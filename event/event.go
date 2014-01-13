package event

import (
	"net"
	"io"
	"crypto/tls"
	"encoding/gob"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/x"
)

type Event struct {
	network, addr, secret string
	conn net.Conn
	decoder *gob.Decoder
	encoder *gob.Encoder
	tlsConfig *tls.Config
}

func New(network, addr, secret string) *Event {
	return &Event{
		network: network,
		addr: addr,
		secret: secret,
	}
}

func (ev *Event) Dial() (err error) {
	if err = ev.Close(); err != nil {
		return
	}
	if ev.conn, err = net.Dial(ev.network, ev.addr); err != nil {
		return
	}
	if ev.tlsConfig != nil {
		ev.conn = tls.Client(ev.conn, ev.tlsConfig)
	}
	ev.decoder = gob.NewDecoder(ev.conn)
	ev.encoder = gob.NewEncoder(ev.conn)
	return
}

func (ev *Event) SetTLS(tlsCert, tlsKey string) (err error) {
	ev.tlsConfig = &tls.Config{}
	ev.tlsConfig.Certificates = make([]tls.Certificate, 1)
	ev.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(tlsCert, tlsKey)
	return
}

func(ev *Event) Close() error {
	if ev.conn != nil {
		return ev.conn.Close()
	}
	return nil
}

func(ev *Event) Serve() (err error) {
	if err = ev.Dial(); err != nil {
		return
	}
	if err = ev.encoder.Encode(x.NewHello(x.Event, ev.secret)); err != nil {
		return
	}
	for {
		var inpack x.Inpack
		if err = ev.decoder.Decode(&inpack); err != nil {
			if e, ok := err.(*net.OpError); ok { // is OpError
				if e.Temporary() { // is Temporary
					continue
				} else { // Reconnect
					if err = ev.Dial(); err != nil {
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
		ev.handle(&inpack)
	}
	return nil
}

func (ev *Event) handle(inpack *x.Inpack) {
	log.Debug(inpack)
}
