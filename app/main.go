package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	cmdOptions := setServerOptions()

	cfg := config.NewConfig(cmdOptions...)
	db := storage.NewStorage()
	server := server.NewServer(cfg, db)

	if err := server.Start(); err != nil {
		log.Println("failed to start the server: %w\n", err)
		os.Exit(1)
	}
}

// Looks for cmdOptions passed by the user to start the server
func setServerOptions() []config.Option {
	var options []config.Option
	cmdOptions := os.Args[1:]

	for i, opt := range cmdOptions {
		if opt == command.Port && len(cmdOptions) > i {
			options = append(options, config.WithPort(cmdOptions[i+1]))
		}
		if opt == command.Replica && len(cmdOptions) > i {
			options = append(options, config.WithSlave(cmdOptions[i+1]))
		}
	}
	return options
}
