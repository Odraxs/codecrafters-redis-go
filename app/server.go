package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	db := storage.NewStorage()

	// Waiting for a connection
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		storageHandler := handler.NewHandler(conn, db)

		go func() {
			err := storageHandler.HandleClient()
			if err != nil {
				log.Printf("something happened with the client %s connection, err: %s\n",
					conn.LocalAddr().String(), err.Error())
			}
		}()
	}
}
