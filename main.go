package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jrallison/go-workers"

	"github.com/gin-gonic/gin"
)

type TriggerRequest struct {
	Device         string    `json:"device"`
	Minutes        float64   `json:"minutes,float64"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time, time_format:"unix"`
}

const TriggerRequestQueue = "requests"

func ProcessTriggerRequest(message *workers.Msg) {
	log.Println("Processed!")
	// do something with your message
	// message.Jid()
	// message.Args() is a wrapper around go-simplejson (http://godoc.org/github.com/bitly/go-simplejson)
}

func main() {
	redisURL := flag.String("redis-address", os.Getenv("REDIS_URL"), "address of Redis instance")
	redisPool := flag.String("redis-pool", os.Getenv("REDIS_POOL"), "number of Redis connections to keep open")
	isWorker := flag.Bool("worker", false, "run as Sidekiq worker node")
	flag.Parse()

	// Connect to Sidekiq
	workers.Configure(map[string]string{
		"server":   *redisURL,
		"pool":     *redisPool,
		"database": "15",
		"process":  "1",
	})

	if *isWorker {
		log.Println("Running as Worker Node")
		runAsWorkerNode()
	} else {
		log.Println("Running as Server Node")
		runAsServerNode()
	}
}

func runAsWorkerNode() {
	// pull messages from "myqueue" with concurrency of 10
	workers.Process(TriggerRequestQueue, ProcessTriggerRequest, 10)

	// stats will be available at http://localhost:8080/stats
	go workers.StatsServer(8080)

	// Blocks until process is told to exit via unix signal
	workers.Run()
}

func runAsServerNode() {
	router := gin.Default()
	router.POST("/create", createTrigger)
	router.Run(":5000")
}

func createTrigger(c *gin.Context) {
	var request TriggerRequest

	err := c.ShouldBind(&request)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	workers.Enqueue(TriggerRequestQueue, request.Device, nil)

	c.Status(http.StatusNoContent)
}
