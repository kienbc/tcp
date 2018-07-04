package tcp

import (
	"net"
	"bufio"
	"errors"
	"fmt"
	"sync"

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
		conn.MarkUnusable()
		return conn.Close()
	}

	select {
	case c.conns <- conn:
		log.Println("========== put: ", conn)
		return nil
	default:
		return conn.Close()
	}
}

func (c *tcpClient) getConns() chan *tcpConn {
	c.mux.RLock()
	conns := c.conns
	c.mux.RUnlock()
	return conns
}

func (c *tcpClient) get() (*tcpConn, error) {
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

func (c *tcpClient) Close() {
	c.mux.Lock()
	conns := c.conns
	c.conns = nil
	c.mux.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	log.Println("Close client")
	for conn := range conns {
		conn.MarkUnusable()
		conn.Close()
	}
}

func (c *tcpClient) Write(payload []byte) (int, error) {
	conn, err := c.get()
	if err != nil {
		//retry
		return 0, err
	}
	defer conn.Close()
	return conn.Write(payload)
}

func (c *tcpClient) WriteString(payload string) (int, error) {
	conn, err := c.get()
	if err != nil {
		//retry
		return 0, err
	}
	defer conn.Close()
	return conn.WriteString(payload)
}

func (c *tcpClient) WriteRune(payload rune) (int, error) {
	conn, err := c.get()
	if err != nil {
		//retry
		return 0, err
	}
	defer conn.Close()
	return conn.WriteRune(payload)
}

func (c *tcpClient) WriteByte(payload byte) error {
	conn, err := c.get()
	if err != nil {
		//retry
		return err
	}
	defer conn.Close()
	return conn.WriteByte(payload)
}