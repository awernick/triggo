package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

func RunAsServerNode(appConfig AppConfig, redisPool *redis.Pool) {
	router := gin.Default()
	router.Use(InjectAppConfig(&appConfig))
	router.Use(InjectRedisPool(redisPool))
	router.POST("/create", CreateTrigger)
	router.Run(":" + appConfig.HTTPPort)
}

func InjectAppConfig(appConfig *AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("AppConfig", appConfig)
		c.Next()
	}
}

func InjectRedisPool(redisPool *redis.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RedisPool", redisPool)
		c.Next()
	}
}

func CreateTrigger(c *gin.Context) {
	appConfig := c.MustGet("AppConfig").(*AppConfig)
	redisPool := c.MustGet("RedisPool").(*redis.Pool)

	var request TriggerRequest
	err := c.ShouldBind(&request)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	if request.SecretKey != (*appConfig).SecretKey {
		log.Printf("Invalid secret key: %s\n", request.SecretKey)
		c.Status(http.StatusUnauthorized)
		return
	}

	_, err = request.ConvertDelayToSeconds()
	if err != nil {
		log.Println("Could not parse delay into seconds")
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	enqueuer := work.NewEnqueuer((*appConfig).Namespace, redisPool)
	_, err = EnqueueRequest(&request, enqueuer)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func EnqueueRequest(request *TriggerRequest, enqueuer *work.Enqueuer) (*work.ScheduledJob, error) {
	delaySecs, err := (*request).ConvertDelayToSeconds()
	if err != nil {
		log.Println("Could not parse request delay time")
		log.Print(err)
		return nil, err
	}

	device := (*request).NormalizedDeviceName()
	triggerType := (*request).NormalizedTriggerType()
	log.Printf("Scheduled {%s} to {%s} in {%d} seconds\n", device, triggerType, delaySecs)

	return enqueuer.EnqueueIn("delay_trigger", delaySecs, work.Q{
		"device":       device,
		"delay":        delaySecs,
		"trigger_type": triggerType,
	})
}
