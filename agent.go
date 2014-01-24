package x

import (
	"crypto/tls"
	"encoding/gob"
	"github.com/mikespook/golib/iptpool"
	"net"
)

type agent struct {
	gw      *Gateway
	decoder *gob.Decoder
	encoder *gob.Encoder
	conn    net.Conn
	authed  bool
	id      string
	role    uint
	ipt     iptpool.ScriptIpt
}

func newAgent(gw *Gateway, conn net.Conn) (a *agent) {
	a = &agent{
		gw:   gw,
		conn: conn,
		ipt:  gw.iptPool.Get(),
	}
	if gw.tlsConfig != nil {
		a.conn = tls.Client(a.conn, gw.tlsConfig)
	}
	a.decoder = gob.NewDecoder(a.conn)
	a.encoder = gob.NewEncoder(a.conn)
	return
}

func (a *agent) Close() (err error) {
	if a.conn != nil {
		err = a.conn.Close()
		a.conn = nil
	}
	a.gw.unregister(a)
	return
}

func (a *agent) Loop() (err error) {
	defer func() {
		a.gw.iptPool.Put(a.ipt)
	}()
	var pack Pack
	for {
		if err := a.decoder.Decode(&pack); err != nil {
			if fatal, err := IsFatal(err); fatal {
				return err
			} else {
				continue
			}
		}
		switch p := pack.Data.(type) {
		case *SignIn:
			if p.Auth(a.gw.secret) {
				a.gw.register(a)
			} else {
				a.Write(Bye("Auth faild!"))
				err = a.Close()
				return
			}
		default:
			go a.handle(&pack)
		}
	}
}

func (a *agent) handle(pack *Pack) {

}

func (a *agent) Write(data interface{}) (err error) {
	var pack Pack
	pack.Data = data
	for {
		if err = a.encoder.Encode(pack); err != nil {
			if fatal, err := IsFatal(err); fatal {
				return err
			} else {
				continue
			}
		}
		break
	}
	return
}
