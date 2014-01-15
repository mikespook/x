package event

import (
	"github.com/mikespook/golib/log"
	"github.com/mikespook/x"
)

type Event struct {
	*x.Agent
}

func New(network, addr, secret string) *Event {
	ev := &Event{}
	ev.Agent = x.NewAgent(network, addr, secret)
	ev.Handler = eventHandler
	return ev
}

func eventHandler(pack x.Pack) {
	log.Debug(pack)
}
