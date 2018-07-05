package tcp

import (
	"bufio"
	"sync"
	"net"
	"log"
)

type tcpConn struct {
	conn     net.Conn
	*bufio.Writer
	client   *tcpClient
	mux      sync.RWMutex
	unusable bool
}

func (this *tcpConn) Close() error {
	this.mux.RLock()
	defer this.mux.RUnlock()

	if this.unusable || this.client.conns == nil {
		if this.conn != nil {
			log.Println("==== close conn: ", this)
			this.Writer = nil
			return this.conn.Close()
		}
		return nil
	}
	return this.client.put(this)
}

func (this *tcpConn) markUnusable() {
	this.mux.Lock()
	this.unusable = true
	this.mux.Unlock()
}
