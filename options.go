package tcp


import (
	"time"
)

type ClientOptions struct {
	Host           string
	Port           int
	ConnectTimeout time.Duration
	MaxIdleConns   int
}

func NewClientOptions() *ClientOptions {
	o := &ClientOptions{
		Host: "",
		Port: 0,
		//AutoReconnect:  true,
		ConnectTimeout: time.Second * 5,
		MaxIdleConns:   10,
	}
	return o
}

func (o *ClientOptions) SetHost(host string) {
	o.Host = host
}

func (o *ClientOptions) SetPort(port int) {
	o.Port = port
}

func (o *ClientOptions) SetConnectTimeout(timeout time.Duration) {
	o.ConnectTimeout = timeout
}

func (o *ClientOptions) SetMaxIdleConns(num int) {
	o.MaxIdleConns = num
}

