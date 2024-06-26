package command

import "fmt"

const (
	Null     = "$-1\r\n"
	Ok       = "+OK\r\n"
	Pong     = "+PONG\r\n"
	Fullsync = "FULLRESYNC"
)

func NewInteger(number int) string {
	return fmt.Sprintf(":%d\r\n", number)
}

func NewString(data string) string {
	return fmt.Sprintf("+%s\r\n", data)
}

func NewBulkString(data string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
}

func NewArray(data []string) string {
	array := fmt.Sprintf("*%d\r\n", len(data))
	for _, d := range data {
		array += NewBulkString(d)
	}

	return array
}

func NewRDBFile(fileContent []byte) string {
	return fmt.Sprintf("$%d\r\n%s", len(fileContent), fileContent)
}
