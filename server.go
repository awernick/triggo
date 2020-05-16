package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
)

func RunAsServerNode() {
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

	if request.Delay() == 0 {
		log.Println("Err: Delay not specified")
		c.Status(http.StatusBadRequest)
		return
	}

	_, err = enqueueRequest(&request, QueueNamespace)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusNoContent)
}

func enqueueRequest(request *TriggerRequest, queueNamespace string) (*work.ScheduledJob, error) {
	enqueuer := work.NewEnqueuer(queueNamespace, redisPool)

	device := (*request).Device
	delay := (*request).Delay()
	triggerKey := (*request).TriggerKey()
	log.Printf("Scheduled {%s} in {%d} seconds\n", triggerKey, delay)

	return enqueuer.EnqueueIn("delay_trigger", (*request).Delay(), work.Q{
		"device":      device,
		"delay":       delay,
		"trigger_key": triggerKey,
	})
}
