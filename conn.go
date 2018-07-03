package tcp

import (
	"sync"
	"net"
	"bufio"
)

type tcpConn struct {
	net.Conn
	writer   *bufio.Writer
	client   *tcpClient
	mux      sync.RWMutex
	unusable bool
}

func (this *tcpConn) Close() error {
	this.mux.RLock()
	defer this.mux.RUnlock()

	if this.unusable {
		if this.Conn != nil {
			return this.Conn.Close()
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

