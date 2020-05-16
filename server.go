package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
)

func RunAsServerNode() {
	router := gin.Default()
	router.POST("/create", createTrigger)
	router.Run(":" + os.Getenv("PORT"))
}

func createTrigger(c *gin.Context) {
	var request TriggerRequest

	err := c.ShouldBind(&request)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	_, err = request.Delay()
	if err != nil {
		log.Println("Could not parse convert delay to seconds")
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	_, err = enqueueRequest(&request, QueueNamespace)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func enqueueRequest(request *TriggerRequest, queueNamespace string) (*work.ScheduledJob, error) {
	enqueuer := work.NewEnqueuer(queueNamespace, redisPool)

	delay, err := (*request).Delay()
	if err != nil {
		log.Println("Could not parse request delay time")
		log.Print(err)
		return nil, err
	}

	device := (*request).NormalizedDeviceName()
	triggerKey := (*request).TriggerKey()
	log.Printf("Scheduled {%s} in {%d} seconds\n", triggerKey, delay)

	return enqueuer.EnqueueIn("delay_trigger", delay, work.Q{
		"device":      device,
		"delay":       delay,
		"trigger_key": triggerKey,
	})
}
