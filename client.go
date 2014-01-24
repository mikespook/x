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
	ErrNilHandler = errors.New("Client.Handler is `nil`.")
	ErrAccessDeny = errors.New("Access Deny")
)

type PackHandler func(Pack)

type Client struct {
	network, addr, secret string
	conn                  net.Conn
	decoder               *gob.Decoder
	encoder               *gob.Encoder
	tlsConfig             *tls.Config

	Id      string
	Handler PackHandler
}

func NewClient(network, addr, secret string) *Client {
	return &Client{
		network: network,
		addr:    addr,
		secret:  secret,
	}
}

func (client *Client) dial() (err error) {
	if err = client.Close(); err != nil {
		return
	}
	if client.conn, err = net.Dial(client.network, client.addr); err != nil {
		return
	}
	if client.tlsConfig != nil {
		client.conn = tls.Client(client.conn, client.tlsConfig)
	}
	client.decoder = gob.NewDecoder(client.conn)
	client.encoder = gob.NewEncoder(client.conn)
	return
}

func (client *Client) SetTLS(tlsCert, tlsKey string) (err error) {
	client.tlsConfig = &tls.Config{}
	client.tlsConfig.Certificates = make([]tls.Certificate, 1)
	client.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(tlsCert, tlsKey)
	return
}

func (client *Client) Close() error {
	if client.conn != nil {
		return client.conn.Close()
	}
	return nil
}

func (client *Client) Loop() (err error) {
	if client.Handler == nil {
		return ErrNilHandler
	}
	if err = client.dial(); err != nil {
		return
	}
	hello := NewSignIn(client.secret)
	if err = client.Write(hello); err != nil {
		return
	}
	for {
		var pack Pack
		if err = client.decoder.Decode(&pack); err != nil {
			if fatal, err := IsFatal(err); fatal {
				return err
			} else {
				continue
			}
		}
		switch p := pack.Data.(type) {
		case Hello:
			client.Id = string(p)
		case Bye:
			err = fmt.Errorf("%s", p)
			return
		default:
			if client.Id == "" {
				err = ErrAccessDeny
				return
			}
			client.Handler(pack)
			// discards
		}
	}
	return
}

func (client *Client) Write(data interface{}) (err error) {
	var pack Pack
	pack.Data = data
	for {
		err = client.encoder.Encode(pack)
		if opErr, ok := err.(*net.OpError); ok { // is OpError
			if opErr.Temporary() { // is Temporary
				continue
			} else { // Reconnect
				if err = client.dial(); err != nil {
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
