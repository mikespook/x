package x

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"strconv"
	"time"
)

const (
	Gateway = iota
	Event
)

func init() {
	gob.Register(&SignIn{})
	gob.Register(Hello("hello"))
	gob.Register(Bye("bye"))
}

type SignIn struct {
	Agent uint
	Token []byte
	Time  time.Time
}

func NewSignIn(agent uint, secret string) *SignIn {
	h := sha1.New()
	now := time.Now()
	io.WriteString(h, strconv.Itoa(int(agent)))
	io.WriteString(h, now.String())
	io.WriteString(h, secret)
	return &SignIn{
		Agent: agent,
		Token: h.Sum(nil),
		Time:  now,
	}
}

func (signIn *SignIn) Auth(secret string) bool {
	h := sha1.New()
	io.WriteString(h, strconv.Itoa(int(signIn.Agent)))
	io.WriteString(h, signIn.Time.String())
	io.WriteString(h, secret)
	return bytes.Compare(h.Sum(nil), signIn.Token) == 0
}

type Hello string

type Bye string

type Pack struct {
	From, To string
	Data     interface{}
}
