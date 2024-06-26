package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type Handler struct {
	db           *storage.Storage
	connection   net.Conn
	cfg          *config.Config
	reader       *bufio.Reader
	writer       *bufio.Writer
	slavesOffset int
	ackSlaves    int
	acksLock     *sync.RWMutex
	acksChan     chan int
}

var commandHandlers = map[string]func(*Handler, *command.Command) error{
	command.Ping:     handlePing,
	command.Echo:     handleEcho,
	command.Get:      handleGet,
	command.Set:      handleSet,
	command.Info:     handleInfo,
	command.Replconf: handleReplconf,
	command.Psync:    handlePsync,
	command.Wait:     handleWait,
	command.Config:   handleConfig,
	command.Keys:     handleKeys,
}

func NewHandler(conn net.Conn, db *storage.Storage, cfg *config.Config, acksChan chan int, locker *sync.RWMutex) *Handler {
	return &Handler{
		db:           db,
		connection:   conn,
		cfg:          cfg,
		reader:       bufio.NewReader(conn),
		writer:       bufio.NewWriter(conn),
		slavesOffset: 0,
		ackSlaves:    0,
		acksLock:     locker,
		acksChan:     acksChan,
	}
}

func (h *Handler) HandleClient() error {
	defer h.connection.Close()

	for {
		userCommand, err := command.NewCommand(h.reader)
		if err != nil {
			return fmt.Errorf("failed to read command, error: %w", err)
		}

		err = h.handleCommand(userCommand)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		// Check if this should only be update for slaves in the tests
		h.cfg.UpdateOffset(userCommand.Size)
		h.writer.Flush()
	}
}

func (h *Handler) Handshake() error {
	h.writer.WriteString(command.NewArray([]string{"PING"}))
	h.writer.Flush()

	response, err := h.reader.ReadString('\n')
	if err != nil || response != command.Pong {
		return fmt.Errorf("incorrect master response")
	}

	h.writer.WriteString(command.NewArray([]string{
		command.Replconf,
		"listening-port",
		h.cfg.Port(),
	}))
	h.writer.Flush()

	response, err = h.reader.ReadString('\n')
	if err != nil || response != command.Ok {
		return fmt.Errorf("incorrect master response")
	}

	h.writer.WriteString(command.NewArray([]string{
		command.Replconf,
		"capa",
		"psync2",
	}))
	h.writer.Flush()

	response, err = h.reader.ReadString('\n')
	if err != nil || response != command.Ok {
		return fmt.Errorf("incorrect master response")
	}

	h.writer.WriteString(command.NewArray([]string{
		command.Psync,
		"?",
		"-1",
	}))
	h.writer.Flush()

	response, err = h.reader.ReadString('\n')
	responseCommand := strings.Split(response, " ")[0]
	if err != nil || responseCommand != "+"+command.Fullsync {
		return fmt.Errorf("incorrect master response")
	}

	// Skip RDB file
	_, err = h.reader.ReadBytes(162)
	if err != nil {
		return err
	}

	return nil
}

// Writes responses to commands send to the master server.
// Slaves only give responses to `REPLCONF GETACK` commands
func (h *Handler) WriteResponse(msg string) {
	if h.cfg.Role() == config.RoleMaster {
		h.writer.WriteString(msg)
	}
}

func (h *Handler) UpdaterSlavesOffset(commandBytes int) {
	h.slavesOffset += commandBytes
}

func (h *Handler) AckSlaves() int {
	h.acksLock.RLock()
	defer h.acksLock.RUnlock()
	return h.ackSlaves
}

func (h *Handler) IncrementAckSlaves() {
	h.acksLock.Lock()
	h.ackSlaves++
	h.acksLock.Unlock()
	h.acksChan <- h.ackSlaves
}

func (h *Handler) SetAckSlaves(val int) {
	h.acksLock.Lock()
	h.ackSlaves = val
	h.acksLock.Unlock()
}

func (h *Handler) handleCommand(userCommand *command.Command) error {
	instruction := strings.ToLower(userCommand.Args[0])
	handler, exist := commandHandlers[instruction]
	if !exist {
		return fmt.Errorf("unknown command: %s", strings.ToUpper(instruction))
	}
	return handler(h, userCommand)
}

func (h *Handler) sendGetAckToSlaves() {
	wg := &sync.WaitGroup{}
	command := command.NewArray([]string{"REPLCONF", "GETACK", "*"})
	for _, slave := range h.cfg.Slaves() {
		wg.Add(1)
		go slave.PropagateCommand(command, wg)
	}
	wg.Wait()
	h.UpdaterSlavesOffset(len([]byte(command)))
}
