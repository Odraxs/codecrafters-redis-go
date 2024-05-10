package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

var (
	portMatch        = regexp.MustCompile(`--port\s+\d+`)
	replicaMatch     = regexp.MustCompile(`--replicaof\s+.+\s\d+`)
	rdbFileDirMatch  = regexp.MustCompile(`--dir\s+[^\s]+`)
	rdbFileNameMatch = regexp.MustCompile(`--dbfilename\s+[^\s]+`)
)

func main() {
	log.Println("Logs from your program will appear here!")

	cmdOptions := setServerOptions()

	cfg := config.NewConfig(cmdOptions...)
	db := storage.NewStorage()
	server := server.NewServer(cfg, db)

	if cfg.Role() == config.RoleSlave {
		if err := server.Handshake(); err != nil {
			log.Printf("failed handshake to %s, error: %s", cfg.ReplicaOf(), err.Error())
			os.Exit(1)
		}
	}

	if err := server.Start(); err != nil {
		log.Printf("failed to start the server: %s\n", err.Error())
		os.Exit(1)
	}
}

// Looks for cmdOptions passed by the user to start the server
func setServerOptions() []config.Option {
	var options []config.Option
	cmdOptions := strings.Join(os.Args[1:], " ")

	if params := portMatch.FindStringSubmatch(cmdOptions); len(params) == 1 {
		port := strings.Split(params[0], " ")[1]
		options = append(options, config.WithPort(port))
	}

	if params := replicaMatch.FindStringSubmatch(cmdOptions); len(params) == 1 {
		data := strings.Split(params[0], " ")[1:]
		master := fmt.Sprintf("%s:%s", data[0], data[1])
		options = append(options, config.WithSlave(master))
	}

	if params := rdbFileDirMatch.FindStringSubmatch(cmdOptions); len(params) == 1 {
		rdbFileDir := strings.Split(params[0], " ")[1]
		options = append(options, config.WithDir(rdbFileDir))
	}

	if params := rdbFileNameMatch.FindStringSubmatch(cmdOptions); len(params) == 1 {
		rdbFileName := strings.Split(params[0], " ")[1]
		options = append(options, config.WithRDBFileName(rdbFileName))
	}

	return options
}
