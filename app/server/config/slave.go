package config

import (
	"bufio"
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

func (sl *Slave) PropagateCommand(args []string, wg *sync.WaitGroup) {
	defer wg.Done()
	command := command.NewArray(args)

	sl.lock.Lock()
	writer := bufio.NewWriter(sl.conn)
	writer.WriteString(command)
	writer.Flush()
	sl.lock.Unlock()
}
