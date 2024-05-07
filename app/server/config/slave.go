package config

import (
	"bufio"
	"net"
	"sync"
)

type Slave struct {
	conn  net.Conn
	lock  *sync.Mutex
}

func NewSlave(conn net.Conn) *Slave {
	return &Slave{
		conn: conn,
		lock: &sync.Mutex{},
	}
}

func (sl *Slave) PropagateCommand(command string, wg *sync.WaitGroup) {
	defer wg.Done()

	sl.lock.Lock()
	writer := bufio.NewWriter(sl.conn)
	writer.WriteString(command)
	writer.Flush()
	sl.lock.Unlock()
}
