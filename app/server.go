package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		// Waiting for a connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go func() {
			err := handleClient(conn)
			if err != nil {
				log.Printf("something happened with the client %s connection, err: %s\n",
					conn.LocalAddr().String(), err.Error())
			}
		}()
	}
}

func handleClient(conn net.Conn) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		userCommand, err := command.NewCommand(reader)
		if err != nil {
			return fmt.Errorf("failed to read command, error: %w", err)
		}

		instruction := strings.ToLower(userCommand.Args[0])
		switch instruction {
		case command.Ping:
			pingMsg := []byte("+PONG\r\n")
			writer.Write(pingMsg)
		case command.Echo:
			if userCommand.Size < 2 {
				return fmt.Errorf("%s command requires an argument", strings.ToUpper(command.Echo))
			}
			arg := userCommand.Args[1]
			writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
		}
		writer.Flush()
	}
}
