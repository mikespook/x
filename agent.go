package x

import (
	"crypto/tls"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	ErrNilHandler = errors.New("Agent.Handler is `nil`.")
	ErrAccessDeny = errors.New("Access Deny")
)

type PackHandler func(Pack)

type Agent struct {
	network, addr, secret string
	conn                  net.Conn
	decoder               *gob.Decoder
	encoder               *gob.Encoder
	tlsConfig             *tls.Config

	Id      string
	Handler PackHandler
}

func NewAgent(network, addr, secret string) *Agent {
	return &Agent{
		network: network,
		addr:    addr,
		secret:  secret,
	}
}

func (agent *Agent) Dial() (err error) {
	if err = agent.Close(); err != nil {
		return
	}
	if agent.conn, err = net.Dial(agent.network, agent.addr); err != nil {
		return
	}
	if agent.tlsConfig != nil {
		agent.conn = tls.Client(agent.conn, agent.tlsConfig)
	}
	agent.decoder = gob.NewDecoder(agent.conn)
	agent.encoder = gob.NewEncoder(agent.conn)
	return
}

func (agent *Agent) SetTLS(tlsCert, tlsKey string) (err error) {
	agent.tlsConfig = &tls.Config{}
	agent.tlsConfig.Certificates = make([]tls.Certificate, 1)
	agent.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(tlsCert, tlsKey)
	return
}

func (agent *Agent) Close() error {
	if agent.conn != nil {
		return agent.conn.Close()
	}
	return nil
}

func (agent *Agent) Serve() (err error) {
	if agent.Handler == nil {
		return ErrNilHandler
	}
	if err = agent.Dial(); err != nil {
		return
	}
	hello := NewSignIn(Event, agent.secret)
	if err = agent.Write(hello); err != nil {
		return
	}
	for {
		var pack Pack
		if err = agent.decoder.Decode(&pack); err != nil {
			if e, ok := err.(*net.OpError); ok { // is OpError
				if e.Temporary() { // is Temporary
					continue
				} else { // Reconnect
					if err = agent.Dial(); err != nil {
						return
					}
				}
			} else { // isn't OpError
				if err == io.EOF {
					err = nil
				}
				return
			}
		}
		switch p := pack.Data.(type) {
		case Hello:
			agent.Id = string(p)
		case Bye:
			err = fmt.Errorf("%s", p)
			return
		default:
			if agent.Id == "" {
				err = ErrAccessDeny
				return
			}
			agent.Handler(pack)
			// discards
		}
	}
	return
}

func (agent *Agent) Write(data interface{}) (err error) {
	var pack Pack
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
