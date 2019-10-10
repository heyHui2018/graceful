# graceful
A golang package that can gracefully restart/stop HTTP service.

###Install
go get -u github.com/heyHui2018/graceful

###Example
```
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/heyHui2018/graceful"
	"github.com/heyHui2018/log"
	"os"
	"syscall"
	"time"
)

func main() {
	router := gin.Default()
	router.GET("/test", func(context *gin.Context) {
		for i := 0; i < 5; i++ {
			log.Info("111111111111111111111")
			time.Sleep(1 * time.Second)
		}
		context.String(200, "ok")
	})
	g := new(graceful.Graceful)
	g.Addr = ":8084"
	g.Handler = router
	g.StopSignalMap = make(map[os.Signal]int)
	g.StopSignalMap[syscall.SIGTERM] = 1
	g.RestartSignalMap = make(map[os.Signal]int)
	g.RestartSignalMap[syscall.SIGINT] = 1
	err := g.Run()
	if err != nil {
		log.Warnf("err = %v", err)
	}
	g.Wg.Wait()
}
```