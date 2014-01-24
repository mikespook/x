package x

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"time"
	"net"
)

func init() {
	gob.Register(&SignIn{})
	gob.Register(Hello("hello"))
	gob.Register(Bye("bye"))
}

type SignIn struct {
	Token []byte
	Time  time.Time
}

func NewSignIn(secret string) *SignIn {
	h := sha1.New()
	now := time.Now()
	io.WriteString(h, now.String())
	io.WriteString(h, secret)
	return &SignIn{
		Token: h.Sum(nil),
		Time:  now,
	}
}

func (signIn *SignIn) Auth(secret string) bool {
	h := sha1.New()
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

func IsFatal(err error) (fatal bool, e error) {
	if opErr, ok := err.(*net.OpError); ok { // is OpError
		fatal = opErr.Temporary() == false
		e = opErr
	} else { // isn't OpError
		fatal = true
		if err != io.EOF {
			e = err
		}
	}
	return
}
