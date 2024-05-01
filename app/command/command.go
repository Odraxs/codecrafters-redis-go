package command

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	Ping = "ping"
	Echo = "echo"
)

const (
	SimpleString = '+'
	BulkString   = '$'
	Arrays       = '*'
)

type Command struct {
	Args []string
	Size uint
}

func NewCommand(reader *bufio.Reader) (*Command, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	var command *Command

	switch line[0] {
	default:
		command = &Command{
			Args: []string{},
		}
	case SimpleString:
		command = &Command{
			Args: []string{line[1:]},
		}
	case BulkString:
		formattedString, err := parseBulkString(reader)

		if err != nil {
			return nil, err
		}
		command = &Command{
			Args: []string{formattedString},
			Size: 1,
		}

	case Arrays:
		args, size, err := parseArray(reader, line)
		if err != nil {
			return nil, err
		}
		command = &Command{
			Args: args,
			Size: uint(size),
		}
	}

	return command, nil
}

func parseBulkString(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimSpace(line)

	size := 0
	fmt.Sscanf(line, "$%d", &size)
	data := make([]byte, size+2)

	_, err = io.ReadFull(reader, data)
	if err != nil {
		return "", err
	}

	return string(data[:size]), nil
}

func parseArray(reader *bufio.Reader, line string) ([]string, int, error) {
	size := 0
	_, err := fmt.Sscanf(line, "*%d", &size)
	if err != nil {
		return nil, 0, err
	}

	arrayArgs := make([]string, size)
	for i := 0; i < size; i++ {
		arg, err := parseBulkString(reader)
		if err != nil {
			return nil, 0, err
		}
		arrayArgs[i] = arg
	}

	return arrayArgs, size, nil
}
