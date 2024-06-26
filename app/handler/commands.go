package handler

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/rdb"
)

func handlePing(h *Handler, _ *command.Command) error {
	pingMsg := command.Pong
	h.WriteResponse(pingMsg)
	return nil
}

func handleEcho(h *Handler, userCommand *command.Command) error {
	if len(userCommand.Args) < 2 {
		return fmt.Errorf("%s command requires an argument", strings.ToUpper(command.Echo))
	}

	arg := userCommand.Args[1]
	h.WriteResponse(command.NewBulkString(arg))
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
	if h.cfg.Role() == config.RoleMaster {
		wg := &sync.WaitGroup{}
		command := command.NewArray(userCommand.Args)
		for _, slave := range h.cfg.Slaves() {
			wg.Add(1)
			go slave.PropagateCommand(command, wg)
		}
		wg.Wait()
		h.UpdaterSlavesOffset(len([]byte(command)))
	}
	h.WriteResponse(command.Ok)

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

func handleConfig(h *Handler, userCommand *command.Command) error {
	config := strings.ToLower(userCommand.Args[1])
	switch config {
	default:
		return fmt.Errorf("%s is an invalid argument", strings.ToUpper(config))
	case command.Get:
		for _, arg := range userCommand.Args[2:] {
			configOf := strings.ToLower(arg)
			if configOf == command.Dir {
				dir := h.cfg.Dir()
				h.WriteResponse(command.NewArray([]string{configOf, dir}))
			}
			if configOf == command.DBfilename {
				fileName := h.cfg.RDBFileName()
				h.WriteResponse(command.NewArray([]string{configOf, fileName}))
			}
		}
	}
	return nil
}

func handleKeys(h *Handler, userCommand *command.Command) error {
	if len(userCommand.Args) > 3 {
		return fmt.Errorf("invalid arguments for the %s command", strings.ToUpper(command.Keys))
	}

	keys := h.db.GetKeys()
	h.WriteResponse(command.NewArray(keys))
	return nil
}

func handleReplconf(h *Handler, userCommand *command.Command) error {
	confOf := strings.ToLower(userCommand.Args[1])
	switch confOf {
	default:
		h.WriteResponse(command.Ok)
	case command.GetAck:
		if h.cfg.Role() == config.RoleMaster {
			info := strings.ToUpper(strings.Join(userCommand.Args, " "))
			return fmt.Errorf("the %s command is only available for salves", info)
		}

		offSet := strconv.Itoa(h.cfg.ReplOffset())
		response := command.NewArray([]string{
			userCommand.Args[0],
			strings.ToUpper(command.Ack),
			offSet,
		},
		)
		h.writer.WriteString(response)
	case command.Ack:
		if h.cfg.Role() == config.RoleSlave {
			info := strings.ToUpper(strings.Join(userCommand.Args, " "))
			return fmt.Errorf("the %s command is only available for master", info)
		}

		// Probably needed for validations latter
		//offSet, _ := strconv.Atoi(userCommand.Args[2])
		h.IncrementAckSlaves()
	}

	return nil
}

func handlePsync(h *Handler, _ *command.Command) error {
	h.writer.WriteString(
		command.NewString(
			fmt.Sprintf("%s %s %d", command.Fullsync, "REPL_ID", 0),
		),
	)

	dbData, err := rdb.GetEmptyRDBContent()
	if err != nil {
		return fmt.Errorf("failed to read rdb file, error: %w", err)
	}
	h.WriteResponse(command.NewRDBFile(dbData))

	slave := config.NewSlave(h.connection)
	h.cfg.AddSlave(slave)
	return nil
}

func handleWait(h *Handler, userCommand *command.Command) error {
	if h.cfg.Role() != config.RoleMaster {
		return fmt.Errorf(
			"the %s command is only allowed for %s servers",
			strings.ToUpper(userCommand.Args[0]),
			strings.ToUpper(config.RoleMaster),
		)
	}
	if len(userCommand.Args) != 3 {
		return fmt.Errorf("the number of argument for %s is incorrect", userCommand.Args[0])
	}

	numReplicas, err := strconv.Atoi(userCommand.Args[1])
	if err != nil {
		return err
	}
	if numReplicas == 0 {
		h.WriteResponse(command.NewInteger(0))
		return nil
	}

	waitTime, err := strconv.Atoi(userCommand.Args[2])
	if err != nil {
		return err
	}

	h.SetAckSlaves(0)
	h.sendGetAckToSlaves()

	for {
		select {
		case ackSlaves := <-h.acksChan:
			// Cheating to prevent race conditions
			// probably should look for a way to manage multiple REPLCONF ACK responses
			if ackSlaves > h.AckSlaves() {
				h.SetAckSlaves(ackSlaves)
			}
			if ackSlaves >= numReplicas {
				h.WriteResponse(command.NewInteger(ackSlaves))
				return nil
			}
		case <-time.After(time.Duration(waitTime) * time.Millisecond):
			if h.AckSlaves() > 0 {
				h.WriteResponse(command.NewInteger(h.AckSlaves()))
			} else {
				h.WriteResponse(command.NewInteger(len(h.cfg.Slaves())))
			}
			return nil
		}
	}
}
