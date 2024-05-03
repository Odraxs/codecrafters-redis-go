package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type Handler struct {
	db         *storage.Storage
	connection net.Conn
	cfg        *config.Config
	reader     *bufio.Reader
	writer     *bufio.Writer
}

var commandHandlers = map[string]func(*Handler, *command.Command) error{
	command.Ping: handlePing,
	command.Echo: handleEcho,
	command.Get:  handleGet,
	command.Set:  handleSet,
	command.Info: handleInfo,
}

func NewHandler(conn net.Conn, db *storage.Storage, cfg *config.Config) *Handler {
	return &Handler{
		db:         db,
		connection: conn,
		cfg:        cfg,
		reader:     bufio.NewReader(conn),
		writer:     bufio.NewWriter(conn),
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

		h.writer.Flush()
	}
}

func (h *Handler) Handshake() error {
	h.writer.WriteString(command.NewArray([]string{"PING"}))
	h.writer.Flush()

	_, err := h.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to get master response, error: %w", err)
	}

	return nil
}

func (h *Handler) handleCommand(userCommand *command.Command) error {
	instruction := strings.ToLower(userCommand.Args[0])
	handler, exist := commandHandlers[instruction]
	if !exist {
		return fmt.Errorf("unknown command: %s", strings.ToUpper(instruction))
	}
	return handler(h, userCommand)
}
