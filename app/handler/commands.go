package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func handlePing(h *Handler, _ *command.Command) error {
	pingMsg := []byte(command.Pong)
	h.writer.Write(pingMsg)
	return nil
}

func handleEcho(h *Handler, userCommand *command.Command) error {
	if len(userCommand.Args) < 2 {
		return fmt.Errorf("%s command requires an argument", strings.ToUpper(command.Echo))
	}

	arg := userCommand.Args[1]
	h.writer.WriteString(command.NewBulkString(arg))
	return nil
}

func handleGet(h *Handler, userCommand *command.Command) error {
	if len(userCommand.Args) < 2 {
		return fmt.Errorf("%s command requires a key argument", strings.ToUpper(command.Get))
	}

	key := userCommand.Args[1]
	value, err := h.db.Get(key)
	if err != nil {
		h.writer.WriteString(command.Null)
	} else {
		h.writer.WriteString(command.NewBulkString(value))
	}
	return nil
}

func handleSet(h *Handler, userCommand *command.Command) error {
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
		var err error
		expTime, err = strconv.Atoi(time)
		if err != nil {
			return fmt.Errorf("the argument after %s should be an integer number", strings.ToUpper(command.Px))
		}
	}

	h.db.Set(key, value, expTime)
	h.writer.WriteString(command.Ok)
	return nil
}

func handleInfo(h *Handler, userCommand *command.Command) error {
	infoOf := strings.ToLower(userCommand.Args[1])
	switch infoOf {
	default:
		return fmt.Errorf("%s is an invalid argument", infoOf)
	case command.Replication:
		info := strings.Join(
			[]string{
				fmt.Sprintf("role:%s", h.cfg.Role()),
				fmt.Sprintf("master_replid:%s", h.cfg.ReplID()),
				fmt.Sprintf("master_repl_offset:%d", h.cfg.ReplOffset()),
			},
			"\n",
		)
		h.writer.WriteString(command.NewBulkString(info))
	}

	return nil
}