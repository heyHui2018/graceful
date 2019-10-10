package graceful

import (
	"os"
	"syscall"
)

func (g *Graceful) defaultStopSignal() {
	g.StopSignalMap = make(map[os.Signal]int)
	g.StopSignalMap[syscall.SIGKILL] = 1
	g.StopSignalMap[syscall.SIGTERM] = 1
}

func (g *Graceful) defaultRestartSignal() {
	g.RestartSignalMap = make(map[os.Signal]int)
	g.RestartSignalMap[syscall.SIGINT] = 1
}
