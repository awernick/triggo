package triggo

import (
	"log"
	"os"
	"os/signal"

	"github.com/gocraft/work"
)

type Context struct{}

func RunAsWorkerNode() {
	pool := work.NewWorkerPool(nil, 10, "my_app_namespace", redisPool)
	pool.Job("", (*Context).ProcessTriggerRequest)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

func (c *Context) ProcessTriggerRequest() {
	log.Println("Testing 123")
}
