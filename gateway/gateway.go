package gateway

import (
	"crypto/tls"
	"github.com/mikespook/golib/log"
	"labix.org/v2/mgo/bson"
	"net"
)

type Gateway struct {
	network, addr, secret string
	listener              net.Listener
	tlsConfig             *tls.Config
	agents                map[string]_Agent
}

func New(netname, addr, secret string) (gw *Gateway) {
	return &Gateway{
		netname: netname,
		addr:    addr,
		secret:  secret,
	}
}

func (gw *Gateway) SetTLS(tlsCert, tlsKey string) (err error) {
	gw.tlsConfig = &tls.Config{}
	gw.tlsConfig.Certificates = make([]tls.Certificate, 1)
	gw.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(tlsCert, tlsKey)
	return
}

func (gw *Gateway) Close() {
	if err := gw.listener.Close(); err != nil {
		log.Error(err)
	}
}

func (gw *Gateway) Serve() (err error) {
	gw.listener, err = net.Listen(gw.netname, gw.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := gw.listener.Accept()
		if err != nil {
			log.Error(err)
			if err.Error() != "use of closed network connection" {
				continue
			}
			break
		}
		if conn == nil {
			break
		}
		go gw.newAgent(conn)
	}
	return nil
}

func (gw *Gateway) newAgent(conn net.Conn) {
	defer log.Messagef("The connection terminated: %s => %s", conn.RemoteAddr(), conn.LocalAddr())
	log.Messagef("New connection established: %s => %s", conn.RemoteAddr(), conn.LocalAddr())
	agent := newAgent(gw, conn)
	defer agent.Close()
	agent.Serve()
	defer agent.Unregister(agent)
}

func (gw *Gateway) register(agent *_Agent) (err error) {
	id := gw.getId()
	if err = agent.Write(x.Hello(id)); err != nil {
		return
	}
	gw.agents[id] = agent
}

func (gw *Gateway) unregister(agent *_Agent) {
	if agent.Id != nil {
		delete(gw.agents, agent.Id)
	}
}

func (gw *Gateway) getId() string {
	return bson.NewObjectId().String()
}
