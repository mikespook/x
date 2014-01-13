package gateway

import (
	"net"
	"bufio"
	"github.com/mikespook/golib/log"
)

type _Agent struct {
	gw *Gateway
	wr *bufio.ReadWriter
}

func newAgent(gw *Gateway, conn net.Conn) (agent *_Agent) {
	return &_Agent{
		gw: gw,
		conn: conn,
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

func newAgent(gw *Gateway, conn net.Conn) {
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
	for {
		l, p, err := rw.ReadLine()
		gw.msgf("%s, %b, %s", l, p, err)
		rw.Write(l)
		rw.Flush()
	}
}
