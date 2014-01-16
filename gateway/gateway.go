package gateway

import (
	"crypto/tls"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/x"
	"labix.org/v2/mgo/bson"
	"net"
	"sync"
)

type Gateway struct {
	network, addr, secret string
	listener              net.Listener
	tlsConfig             *tls.Config
	agents                map[string]*_Agent
	sync.RWMutex
}

func New(network, addr, secret string) (gw *Gateway) {
	return &Gateway{
		network: network,
		addr:    addr,
		secret:  secret,
		agents:  make(map[string]*_Agent, 16),
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
	gw.listener, err = net.Listen(gw.network, gw.addr)
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
	defer gw.unregister(agent)
}

func (gw *Gateway) register(role uint, agent *_Agent) (err error) {
	gw.Lock()
	defer gw.Unlock()
	agent.id = gw.getId()
	if err = agent.Write(x.Hello(agent.id)); err != nil {
		return
	}
	agent.role = role
	gw.agents[agent.id] = agent
	log.Messagef("The agent registered: %x", agent.id)
	return
}

func (gw *Gateway) unregister(agent *_Agent) {
	gw.Lock()
	defer gw.Unlock()
	if agent.id != "" {
		delete(gw.agents, agent.id)
		log.Messagef("The agent unregistered: %x", agent.id)
	}
}

func (gw *Gateway) getId() string {
	return string(bson.NewObjectId())
}
