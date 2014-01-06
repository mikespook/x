package gateway

import (
	"net"
	"bufio"
	"github.com/mikespook/golib/log"
)

type Gateway struct {
	netname, addr string
	listener net.Listener
	logger *log.Logger
}

func New(netname, addr string) (gw *Gateway) {
	return &Gateway{
		netname: netname,
		addr: addr,
	}
}

func(gw *Gateway) msgf(format string, msg ... interface{}) {
	if gw.logger != nil {
		gw.logger.Messagef(format, msg ...)
	}
}

func(gw *Gateway) err(err error) {
	if gw.logger != nil {
		gw.logger.Error(err)
	}
}

func(gw *Gateway) SetLogger(logger *log.Logger) {
	gw.logger = logger
}

func(gw *Gateway) Close() {
	if err := gw.listener.Close(); err != nil {
		gw.err(err)
	}
}

func(gw *Gateway) Serve() (err error) {
	gw.listener, err = net.Listen(gw.netname, gw.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := gw.listener.Accept()
		if err != nil {
			gw.err(err)
			if err.Error() != "use of closed network connection" {
				continue
			}
			break
		}
		if conn == nil {
			break
		}
		go gw.newClient(conn)
	}
	return nil
}

func (gw *Gateway) newClient(conn net.Conn) {
	defer func() {
		gw.msgf("The connection terminated: %s => %s", conn.RemoteAddr(), conn.LocalAddr())
		if err := conn.Close(); err != nil {
			gw.err(err)
		}
	}()
	gw.msgf("New connection established: %s => %s", conn.RemoteAddr(), conn.LocalAddr())
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	rw := bufio.NewReadWriter(r, w)
	rw.WriteString("Hello")
	rw.Flush()
}
