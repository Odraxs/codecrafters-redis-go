package command

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	Ping     = "ping"
	Echo     = "echo"
	Set      = "set"
	Get      = "get"
	Info     = "info"
	Replconf = "replconf"
	Psync    = "psync"
	Px       = "px"
)

const (
	Replication = "replication"
	GetAck      = "getack"
	Ack         = "ack"
)

const (
	SimpleString = '+'
	BulkString   = '$'
	Arrays       = '*'
)

type Command struct {
	Args []string
	Size int
}

func NewCommand(reader *bufio.Reader) (*Command, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	command := &Command{
		Size: len([]byte(line)),
	}
	line = strings.TrimSpace(line)

	switch line[0] {
	default:
		command.Args = []string{}

	case SimpleString:
		command.Args = []string{line[1:]}

	case BulkString:
		formattedString, bytes, err := parseBulkString(reader)
		if err != nil {
			return nil, err
		}

		command.Args = []string{formattedString}
		command.Size += bytes

	case Arrays:
		args, bytes, err := parseArray(reader, line)
		if err != nil {
			return nil, err
		}

		command.Args = args
		command.Size += bytes
	}

	return command, nil
}

func parseBulkString(reader *bufio.Reader) (string, int, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", 0, err
	}

	bytes := len([]byte(line))
	line = strings.TrimSpace(line)

	size := 0
	fmt.Sscanf(line, "$%d", &size)
	data := make([]byte, size+2)

	commandBytes, err := io.ReadFull(reader, data)
	bytes += commandBytes
	if err != nil {
		return "", 0, err
	}

	return string(data[:size]), bytes, nil
}

func parseArray(reader *bufio.Reader, line string) ([]string, int, error) {
	size := 0
	_, err := fmt.Sscanf(line, "*%d", &size)
	if err != nil {
		return nil, 0, err
	}

	arrayArgs := make([]string, size)
	totalBytes := 0
	for i := 0; i < size; i++ {
		arg, bytes, err := parseBulkString(reader)
		if err != nil {
			return nil, 0, err
		}
		totalBytes += bytes
		arrayArgs[i] = arg
	}

	return arrayArgs, totalBytes, nil
}
