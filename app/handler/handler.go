package handler

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type Handler struct {
	db         *storage.Storage
	connection net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
}

func NewHandler(conn net.Conn, db *storage.Storage) *Handler {
	return &Handler{
		db:         db,
		connection: conn,
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

		instruction := strings.ToLower(userCommand.Args[0])
		switch instruction {
		case command.Ping:
			pingMsg := []byte("+PONG\r\n")
			h.writer.Write(pingMsg)

		case command.Echo:
			if len(userCommand.Args) < 2 {
				return fmt.Errorf("%s command requires an argument", strings.ToUpper(command.Echo))
			}

			arg := userCommand.Args[1]
			h.writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))

		case command.Get:
			if len(userCommand.Args) < 2 {
				return fmt.Errorf("%s command requires a key argument", strings.ToUpper(command.Get))
			}

			key := userCommand.Args[1]
			value, err := h.db.Get(key)
			if err != nil {
				h.writer.WriteString("$-1\r\n")
			} else {
				h.writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
			}

		case command.Set:
			if len(userCommand.Args) < 3 {
				return fmt.Errorf("%s command requires at least key and value arguments", strings.ToUpper(command.Set))
			}
			if len(userCommand.Args) > 5 {
				return fmt.Errorf("invalid command arguments")
			}

			key, value := userCommand.Args[1], userCommand.Args[2]
			expTime := 0

			if len(userCommand.Args) == 5 {
				expInstruction := strings.ToLower(userCommand.Args[3])
				if expInstruction != command.Px {
					return fmt.Errorf("the command %s only allows the %s as a complimenting command",
						strings.ToUpper(command.Set), strings.ToUpper(command.Px))
				}
				time := userCommand.Args[4]
				expTime, err = strconv.Atoi(time)
				if err != nil {
					return fmt.Errorf("the argument after %s should be an integer number", strings.ToUpper(command.Px))
				}
			}

			h.db.Set(key, value, expTime)
			h.writer.WriteString("+OK\r\n")
		}
		h.writer.Flush()
	}
}
