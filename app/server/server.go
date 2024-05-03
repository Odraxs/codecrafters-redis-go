package server

import (
	"fmt"
	"log"
	"net"

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

	// Waiting for a connection
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}

		connHandler := handler.NewHandler(conn, s.db, s.cfg)

		go func() {
			err := connHandler.HandleClient()
			if err != nil {
				log.Printf("something happened with the client %s connection, err: %s\n",
					conn.LocalAddr().String(), err.Error())
			}
		}()
	}
}

func (s *Server) Handshake() error {
	conn, err := net.Dial("tcp", s.cfg.ReplicaOf())
	if err != nil {
		return fmt.Errorf("failed to dial with master error: %w", err)
	}
	defer conn.Close()

	connHandler := handler.NewHandler(conn, s.db, s.cfg)
	if err := connHandler.Handshake(); err != nil {
		return fmt.Errorf("failed to handshake, error: %w", err)
	}

	return nil
}
