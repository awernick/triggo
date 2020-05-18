package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

func RunAsServerNode(appConfig AppConfig, redisPool *redis.Pool) {
	var err error

	// Load Device Mappings
	deviceMapper := DeviceMapper{}
	err = deviceMapper.LoadMappings()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Mappings: %s", deviceMapper.mappings)
	}

	err = appConfig.ValidateHTTPPort()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(InjectAppConfig(&appConfig))
	router.Use(InjectRedisPool(redisPool))
	router.Use(InjectDeviceMapper(&deviceMapper))
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

func InjectDeviceMapper(deviceMapper *DeviceMapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DeviceMapper", deviceMapper)
		c.Next()
	}
}

func CreateTrigger(c *gin.Context) {
	appConfig := c.MustGet("AppConfig").(*AppConfig)
	redisPool := c.MustGet("RedisPool").(*redis.Pool)
	deviceMapper := c.MustGet("DeviceMapper").(*DeviceMapper)

	var request TriggerRequest
	err := c.ShouldBind(&request)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusBadRequest)
		return
	}

	if request.SecretKey != appConfig.SecretKey {
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

	supportedDeviceName := deviceMapper.MapToSupportedDevice(request.NormalizedDeviceName())
	if supportedDeviceName != "" {
		log.Printf("Mapping {%s} to {%s}\n", request.DeviceName, supportedDeviceName)
		request.DeviceName = supportedDeviceName
	} else {
		log.Printf("Could not map device name: {%s}", request.DeviceName)
		log.Printf("Could not map device name: {%s}", request.NormalizedDeviceName())
	}

	enqueuer := work.NewEnqueuer(appConfig.Namespace, redisPool)
	_, err = EnqueueRequest(&request, enqueuer)
	if err != nil {
		log.Print(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func EnqueueRequest(request *TriggerRequest, enqueuer *work.Enqueuer) (*work.ScheduledJob, error) {
	delaySecs, err := request.ConvertDelayToSeconds()
	if err != nil {
		log.Println("Could not parse request delay time")
		log.Print(err)
		return nil, err
	}

	device := request.NormalizedDeviceName()
	triggerType := request.NormalizedTriggerType()
	triggerKey := request.TriggerKey()
	log.Printf("[%s]: Scheduled {%s} to {%s} in {%d} seconds\n", triggerKey, device, triggerType, delaySecs)

	return enqueuer.EnqueueIn("delay_trigger", delaySecs, work.Q{
		"device":      device,
		"delay":       delaySecs,
		"trigger_key": triggerKey,
	})
}
