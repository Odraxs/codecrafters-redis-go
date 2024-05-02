package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	var options []config.Option
	port := findPort()
	if port != "" {
		options = append(options, config.WithPort(port))
	}
	cfg := config.NewConfig(options...)

	l, err := net.Listen("tcp", "0.0.0.0:"+cfg.Port())
	if err != nil {
		fmt.Printf("Failed to bind to port %s\n", cfg.Port())
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

func findPort() string {
	args := os.Args[1:]

	for i, arg := range args {
		if arg == command.Port && len(args) > i {
			return args[i+1]
		}
	}
	return ""
}
