package tcp

import (
	"sync"
	"errors"
	"bufio"
	"net"
	"fmt"
	"log"
)

type tcpClient struct {
	opts  *ClientOptions
	conns chan *tcpConn
	mux   sync.RWMutex
}

func NewClient(opts *ClientOptions) (*tcpClient, error) {
	if opts.Host == "" {
		return nil, errors.New("Host is invalid")
	}

	if opts.Port <= 0 {
		return nil, errors.New("Port is invalid")
	}

	if opts.MaxIdleConns <= 0 {
		return nil, errors.New("invalid MaxIdleConns")
	}

	return &tcpClient{
		opts: opts,
	}, nil
}

func (c *tcpClient) createConn() (*tcpConn, error) {
	var (
		err  error
		conn net.Conn
	)

	address := fmt.Sprintf("%s:%d", c.opts.Host, c.opts.Port)
	conn, err = net.DialTimeout("tcp", address, c.opts.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	return &tcpConn{
		conn:   conn,
		Writer: bufio.NewWriter(conn),
		client: c,
	}, nil
}

func (c *tcpClient) Connect() error {

	c.conns = make(chan *tcpConn, c.opts.MaxIdleConns)

	for i := 0; i < c.opts.MaxIdleConns; i++ {
		conn, err := c.createConn()
		if err != nil {
			c.Close()
			fmt.Errorf("Factory is not able to create connection to fill the pool: %s", err)
			return err
		}
		c.conns <- conn
	}
	return nil
}

func (c *tcpClient) Close() {
	c.mux.Lock()
	conns := c.conns
	c.conns = nil
	c.mux.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *tcpClient) Conn() (*tcpConn, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, errors.New("Client is closed")
	}
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, errors.New("Client is closed")
		}
		return conn, nil
	default:
		conn, err := c.createConn()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (c *tcpClient) put(conn *tcpConn) error {
	if conn == nil {
		return errors.New("Connection is nil. Rejecting")
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	select {
	case c.conns <- conn:
		log.Println("==== put conn: ", conn)
		return nil
	default:
		log.Println("==== put close conn: ", conn)
		return conn.conn.Close()
	}
}

func (c *tcpClient) getConns() chan *tcpConn {
	c.mux.RLock()
	conns := c.conns
	c.mux.RUnlock()
	return conns
}
