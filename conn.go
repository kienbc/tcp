package tcp

import (
	"sync"
	"net"
	"bufio"
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

	if this.unusable {
		if this.conn != nil {
			this.Writer = nil
			return this.conn.Close()
		}
		return nil
	}
	return this.client.put(this)
}

func (this *tcpConn) MarkUnusable() {
	this.mux.Lock()
	this.unusable = true
	this.mux.Unlock()
}
