package config

import (
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

type Slave struct {
	conn net.Conn
	lock *sync.Mutex
}

func NewSlave(conn net.Conn) *Slave {
	return &Slave{
		conn: conn,
		lock: &sync.Mutex{},
	}
}

func (si *Slave) PropagateCommand(args []string) {
	command := command.NewArray(args)
	si.lock.Lock()
	si.conn.Write([]byte(command))
	si.lock.Unlock()
}
