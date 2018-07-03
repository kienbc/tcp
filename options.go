package tcp

import (
	"time"
)

type ClientOptions struct {
	Host           string
	Port           int
	AutoReconnect  bool
	ConnectTimeout time.Duration

	// if MaxIdleConns <= 0 && MaxConns <= 0, remove pool from connection
	MaxIdleConns int
	MaxConns     int
}

func NewClientOptions() *ClientOptions {
	o := &ClientOptions{
		Host:           "",
		Port:           0,
		AutoReconnect:  true,
		ConnectTimeout: time.Second * 5,
		MaxIdleConns:   5,
		MaxConns:       20,
	}
	return o
}

func (o *ClientOptions) SetHost(host string) {
	o.Host = host
}

func (o *ClientOptions) SetPort(port int) {
	o.Port = port
}

func (o *ClientOptions) SetAutoReConnect(reconnect bool) {
	o.AutoReconnect = reconnect
}

func (o *ClientOptions) SetConnectTimeout(timeout time.Duration) {
	o.ConnectTimeout = timeout
}

func (o *ClientOptions) SetMaxIdleConns(num int) {
	o.MaxIdleConns = num
}

func (o *ClientOptions) SetMaxConns(num int)  {
	o.MaxConns = num
}