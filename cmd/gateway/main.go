package main

import (
	"flag"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/golib/pid"
	"github.com/mikespook/golib/signal"
	"github.com/mikespook/x/gateway"
	"os"
	"syscall"
	"net"
)

const (
	GITLAB = "gitlab"
	GITHUB = "github"
)

var (
	addr       = flag.String("addr", "127.0.0.1:8081", "Address of socket")
	tlsCert       = flag.String("tls-cert", "", "TLS cert file")
	tlsKey        = flag.String("tls-key", "", "TLS key file")
	pf         = flag.String("pid", "", "PID file")
)

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
	log.Flag()
}

func main() {
	log.Messagef("Starting: addr=%q", *addr)
	if *pf != "" {
		if p, err := pid.New(*pf); err != nil {
			log.Error(err)
		} else {
			defer func() {
				if err := p.Close(); err != nil {
					log.Error(err)
				}
			}()
			log.Messagef("PID: %d file=%q", p.Pid, *pf)
		}
	}
	defer func() {
		log.Message("Exited!")
		log.WaitClosing()
	}()
	// Init Gateway
	gw := gateway.New(*addr)
	if *tlsCert != "" && *tlsKey != "" {
		// Use TLS connection
		log.Messagef("Using TLS: cert=%q; key=%q", *cert, *key)
		if err := gw.SetTLS(*tlsCert, *tlsKey); err != nil {
			log.Error(err)
			return
		}
	}
	defer gw.Close()
	// Goroutine a Gateway logical
	go func() {
		if err := gw.Serve(); err != nil {
			if _, ok := err.(*net.OpError); !ok {
				log.Error(err)
				signal.Send(os.Getpid(), os.Interrupt)
			}
		}
	}()
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool { return true })
	sh.Bind(syscall.SIGUSR1, func() bool {
		// Shutdown Gateway
		gw.Close()
		if err:= restart(); err != nil {
			log.Error(err)
		}
		return true
	})
	sh.Loop()
}

func restart() error {
	var attr os.ProcAttr
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	attr.Sys = &syscall.SysProcAttr{}
	_, err := os.StartProcess(os.Args[0], os.Args, &attr)
	return err
}
