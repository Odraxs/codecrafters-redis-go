package main

import (
	"fmt"
	// Uncomment this block to pass the first stage
	"net"
	"os"
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

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 128)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Failed to read connection data, error: %v\n", err.Error())
			os.Exit(1)
		}

		pingMsg := []byte("+PONG\r\n")
		_, err = conn.Write(pingMsg)
		if err != nil {
			fmt.Printf("Failed to write ping message, error: %v\n", err.Error())
			os.Exit(1)
		}
	}
}
