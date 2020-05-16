package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gocraft/work"
)

type Context struct{}

func RunAsWorkerNode() {
	pool := work.NewWorkerPool(Context{}, 10, QueueNamespace, redisPool)
	pool.Job("delay_trigger", (*Context).ProcessTriggerRequest)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

func (c *Context) ProcessTriggerRequest(job *work.Job) error {
	device := job.ArgString("device")
	delay := job.ArgInt64("delay")

	if err := job.ArgError(); err != nil {
		log.Println(err)
		return err
	}

	log.Println(fmt.Sprintf("Device: %s", device))
	log.Println(fmt.Sprintf("Delay: %d", delay))

	// do something with your message // message.Jid()
	// message.Args() is a wrapper around go-simplejson (http://godoc.org/github.com/bitly/go-simplejson)
	return nil
}
