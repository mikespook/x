package x

import (
	"crypto/tls"
	"github.com/mikespook/golib/log"
	"labix.org/v2/mgo/bson"
	"net"
	"sync"
	"sort"
)

type Gateway struct {
	network, addr, secret string
	listener              net.Listener
	tlsConfig             *tls.Config
	agents                map[string]*agent
	groups                map[string]sort.StringSlice
	mutex sync.RWMutex
}

func New(network, addr, secret string) (gw *Gateway) {
	return &Gateway{
		network: network,
		addr:    addr,
		secret:  secret,
		agents:  make(map[string]*agent, 16),
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

func (gw *Gateway) Loop() (err error) {
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
	if err := agent.Loop(); err != nil {
		log.Error(err)
	}
	defer gw.unregister(agent)
}

func (gw *Gateway) register(a *agent) (err error) {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	a.id = gw.getId()
	if err = a.Write(Hello(a.id)); err != nil {
		return
	}
	gw.agents[a.id] = a
	log.Messagef("The agent registered: %x", a.id)
	return
}

func (gw *Gateway) unregister(a *agent) {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	if a.id != "" {
		delete(gw.agents, a.id)
		log.Messagef("The agent unregistered: %x", a.id)
	}
}

func (gw *Gateway) getId() string {
	return string(bson.NewObjectId())
}
