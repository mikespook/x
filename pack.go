package x

import (
	"io"
	"strconv"
	"crypto/sha1"
	"time"
	"bytes"
)

const (
	Gateway = iota
	Event
)

type Hello struct {
	Agent uint
	Token []byte
	Time time.Time
}

func NewHello(agent uint, secret string) *Hello {
	h := sha1.New()
	now := time.Now()
	io.WriteString(h, strconv.Itoa(int(agent)))
	io.WriteString(h, now.String())
	io.WriteString(h, secret)
	return &Hello{
		Agent: agent,
		Token: h.Sum(nil),
		Time: now,
	}
}

func (hello *Hello) Auth(secret string) bool {
	h := sha1.New()
	io.WriteString(h, strconv.Itoa(int(hello.Agent)))
	io.WriteString(h, hello.Time.String())
	io.WriteString(h, secret)
	return bytes.Compare(h.Sum(nil), hello.Token) == 0
}

type Pack interface {
	Bye() bool
}

type Inpack struct {
	Type uint
}

type Outpack struct {
	Type uint
}
