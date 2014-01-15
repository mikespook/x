package gateway

import (
	"crypto/tls"
	"encoding/gob"
	"errors"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/x"
	"io"
	"net"
)

type _Agent struct {
	gw      *Gateway
	decoder *gob.Decoder
	encoder *gob.Encoder
	conn    net.Conn
	authed  bool
}

func newAgent(gw *Gateway, conn net.Conn) (agent *_Agent) {
	agent = &_Agent{
		gw:   gw,
		conn: conn,
	}
	if gw.tlsConfig != nil {
		agent.conn = tls.Client(agent.conn, gw.tlsConfig)
	}
	agent.decoder = gob.NewDecoder(agent.conn)
	agent.encoder = gob.NewEncoder(agent.conn)
	return
}

func (agent *_Agent) Close() {
	if err := agent.conn.Close(); err != nil {
		log.Error(err)
	}
}

func (agent *_Agent) Serve() {
	var pack x.Pack
	for {
		if err := agent.decoder.Decode(&pack); err != nil {
			if _, ok := err.(*net.OpError); !ok && err != io.EOF {
				log.Error(err)
			}
			break
		}
		switch pack := a.Data.(type) {
		case *x.SignIn:
			if pack.Auth(agent.gw.secret) {
				agent.gw.Registe(agent)
			} else {
				agent.Write(x.Bye("Auth faild!"))
				return
			}
		default:
			go agent.handle(pack)
		}
	}
}

func (agent *_Agent) handle(pack x.Pack) {
	log.Debug(pack)
}

func (agent *_Agent) Write(data interface{}) (err error) {
	var pack x.Pack
	pack.Data = data
	for {
		err = agent.encoder.Encode(pack)
		if opErr, ok := err.(*net.OpError); ok { // is OpError
			if opErr.Temporary() { // is Temporary
				continue
			} else { // Reconnect
				if err = agent.Dial(); err != nil {
					return
				}
			}
			continue
		} else { // isn't OpError
			if err == io.EOF {
				err = nil
			}
			return
		}
		break
	}
	return
}
