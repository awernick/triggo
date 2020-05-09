package triggo

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/jrallison/go-workers"
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

	_, err = enqueueRequest(&request, "trigger_request")
	if err != nil {
		log.Print(err)
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusNoContent)
}

func enqueueRequest(request *TriggerRequest, queueNamespace string) (*work.Job, error) {
	enqueuer := work.NewEnqueuer(queueNamespace, redisPool)
	return enqueuer.Enqueue("trigger_request", work.Q{"device": (*request).Device, "delay": (*request).Delay})
}

func ProcessTriggerRequest(message *workers.Msg) {
	log.Println("Processed!")
	// do something with your message // message.Jid()
	// message.Args() is a wrapper around go-simplejson (http://godoc.org/github.com/bitly/go-simplejson)
}
