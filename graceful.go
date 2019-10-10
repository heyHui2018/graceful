package graceful

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"sync"
	"time"
)

var graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")

type Graceful struct {
	Addr             string
	Handler          http.Handler
	server           *http.Server
	listener         net.Listener
	Timeout          time.Duration // unit:second
	StopSignalMap    map[os.Signal]int
	RestartSignalMap map[os.Signal]int
	signals          []os.Signal
	Wg               *sync.WaitGroup
}

func (g *Graceful) check() error {
	if g.Addr == "" {
		return errors.New("addr has not been set")
	}
	if g.Handler == nil {
		g.Handler = http.DefaultServeMux
	}
	g.server = &http.Server{
		Addr:         g.Addr,
		Handler:      g.Handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if g.Timeout <= 0 {
		g.Timeout = 5 * time.Second
	}
	g.Wg = new(sync.WaitGroup)
	if len(g.StopSignalMap) == 0 {
		g.defaultStopSignal()
	}
	if len(g.RestartSignalMap) == 0 {
		g.defaultRestartSignal()
	}
	signals := make([]os.Signal, 0, len(g.StopSignalMap)+len(g.RestartSignalMap))
	for k := range g.StopSignalMap {
		signals = append(signals, k)
	}
	for k := range g.RestartSignalMap {
		signals = append(signals, k)
	}
}

func (g *Graceful) Run() error {
	flag.Parse()
	err := g.check()
	if err != nil {
		return err
	}
	if *graceful {
		f := os.NewFile(3, "")
		g.listener, err = net.FileListener(f)
	} else {
		g.listener, err = net.Listen("tcp", g.Addr)
	}
	if err != nil {
		return err
	}
	go g.signalHandle()
	err = g.server.Serve(g.listener)
	if err != nil {
		return err
	}
	return nil
}

func (g *Graceful) signalHandle() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, g.signals...)
	for {
		select {
		case s := <-signalChan:
			if _, ok := g.StopSignalMap[s]; ok {
				g.Stop()
			} else if _, ok := g.RestartSignalMap[s]; ok {
				g.ReStart()
			}
		}
	}
}

func (g *Graceful) Stop() error {
	g.Wg.Add(1)
	cxt, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := g.server.Shutdown(cxt)
	g.Wg.Done()
	if err != nil {
		return err
	}
}

func (g *Graceful) ReStart() error {
	listener, ok := g.listener.(*net.TCPListener)
	if !ok {
		return errors.New(fmt.Sprintf("listener's type is not *net.TCPListener,type = %v", reflect.TypeOf(g.listener).Name()))
	}
	file, err := listener.File()
	if err != nil {
		return err
	}
	args := os.Args
	if !*graceful {
		args = append(args, "-graceful")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{file}
	err = cmd.Start()
	if err != nil {
		return err
	}
	g.Stop()
	return nil
}
