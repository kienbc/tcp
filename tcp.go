package tcp

import (
	"net"
	"bufio"
	"errors"
	"fmt"
	"sync"
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

	if opts.MaxIdleConns <= 0 || opts.MaxConns <= 0 || opts.MaxIdleConns > opts.MaxConns {
		return nil, errors.New("invalid MaxConns or MaxIdleConns")
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
		Conn:   conn,
		writer: bufio.NewWriter(conn),
		client: c,
	}, nil
}

func (c *tcpClient) Connect() error {

	c.conns = make(chan *tcpConn, c.opts.MaxConns)

	for i := 0; i < c.opts.MaxIdleConns; i++ {
		conn, err := c.createConn()
		if err != nil {
			c.Close()
			fmt.Errorf("Factory is not able to create connection to fill the pool: %s", err)
		}
		c.conns <- conn
	}
	return nil
}

func (c *tcpClient) reConnect() error {
	return nil
}

func (c *tcpClient) put(conn *tcpConn) error {
	if conn == nil {
		return errors.New("Connection is nil. Rejecting")
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.conns == nil {
		return conn.Close()
	}

	select {
	case c.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (c *tcpClient) Close()  {
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
