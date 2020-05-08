package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TriggerRequest struct {
	Device         string    `json:"device"`
	Minutes        float64   `json:"minutes,float64"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time, time_format:"unix"`
}

func main() {
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

	c.Status(http.StatusNoContent)
}
