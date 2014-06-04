package tcp

import (
	"encoding/gob"
	"errors"
	goNet "net"
	"time"
)

type Tcp struct {
	listener *goNet.TCPListener
}

func NewTcp() *Tcp {
	return &Tcp{}
}

func (tcp *Tcp) StartServer(ip, port string, fn func(conn goNet.Conn)) (err error) {
	if nil != tcp.listener {
		return errors.New("already started a TCP server")
	}
	tcpAddr, err := goNet.ResolveTCPAddr("tcp4", ip+":"+port)
	if nil != err {
		return
	}
	listener, err := goNet.ListenTCP("tcp4", tcpAddr)
	if nil != err {
		return
	}
	tcp.listener = listener
	go func(listener *goNet.TCPListener) {
		for {
			conn, err := listener.Accept()
			if nil != err {
				continue
			}
			go func(conn goNet.Conn) {
				defer conn.Close()
				fn(conn)
			}(conn)
		}
	}(listener)
	return
}

func (tcp *Tcp) connect(ip, port string) (ret *goNet.TCPConn, err error) {
	conn, err := goNet.DialTimeout("tcp4", ip+":"+port, 3*time.Second)
	if nil != err {
		return nil, err
	}
	ret, ok := conn.(*goNet.TCPConn)
	if !ok {
		return nil, errors.New("connection is not TCPConn")
	}
	return
}

func (tcp *Tcp) SendMsg(ip, port string, arg interface{}, callback func(conn goNet.Conn) error) error {
	conn, err := tcp.connect(ip, port)
	if nil != err {
		return err
	}
	defer conn.Close()
	err = gob.NewEncoder(conn).Encode(arg)
	if nil != err {
		return err
	}
	err = callback(conn)
	if nil != err {
		return err
	}
	return nil
}
