package main

import (
	"flag"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/golib/pid"
	"github.com/mikespook/golib/signal"
	"github.com/mikespook/x/gateway"
	"os"
	"syscall"
)

var (
	netname = flag.String("net", "tcp", "Network interface")
	addr    = flag.String("addr", "127.0.0.1:8081", "Address of gateway")
	tlsCert = flag.String("tls-cert", "", "TLS cert file")
	tlsKey  = flag.String("tls-key", "", "TLS key file")
	pf      = flag.String("pid", "", "PID file")
	secret  = flag.String("secret", "", "Secret key")
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
	// Init Gateway
	gw := gateway.New(*netname, *addr, *secret)
	if *tlsCert != "" && *tlsKey != "" {
		gw.SetTLS(*tlsCert, *tlsKey)
	}
	// Goroutine a Gateway logical
	go func() {
		if err := gw.Serve(); err != nil {
			log.Error(err)
			signal.Send(os.Getpid(), os.Interrupt)
		}
	}()
	defer func() {
		log.Message("Exited!")
		log.WaitClosing()
	}()
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool {
		return true
	})
	sh.Bind(syscall.SIGUSR1, func() bool {
		// Shutdown Gateway
		gw.Close()
		if err := restart(); err != nil {
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
