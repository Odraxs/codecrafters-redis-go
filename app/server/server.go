package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/server/config"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type Server struct {
	cfg *config.Config
	db  *storage.Storage
}

func NewServer(cfg *config.Config, db *storage.Storage) *Server {
	return &Server{
		cfg: cfg,
		db:  db,
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", "0.0.0.0:"+s.cfg.Port())
	if err != nil {
		return fmt.Errorf("failed to bind to port %s", s.cfg.Port())
	}
	defer l.Close()

	acksChan := make(chan int, 10)
	locker := &sync.RWMutex{}
	// Waiting for a connection
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}

		connHandler := handler.NewHandler(conn, s.db, s.cfg, acksChan, locker)

		go s.serveConnection(connHandler)
	}
}

func (s *Server) Handshake() error {
	conn, err := net.Dial("tcp", s.cfg.ReplicaOf())
	if err != nil {
		return fmt.Errorf("failed to dial with master error: %w", err)
	}

	acksChan := make(chan int, 10)
	connHandler := handler.NewHandler(conn, s.db, s.cfg, acksChan, &sync.RWMutex{})
	if err := connHandler.Handshake(); err != nil {
		return fmt.Errorf("failed to handshake, error: %w", err)
	}

	go s.serveConnection(connHandler)

	return nil
}

func (s *Server) serveConnection(connHandler *handler.Handler) {
	err := connHandler.HandleClient()
	if err != nil {
		log.Printf("something happened with the client connection, err: %s\n", err.Error())
	}
}
