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
	router.POST("/create", CreateDelayedTrigger)
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

func CreateDelayedTrigger(c *gin.Context) {
	appConfig := c.MustGet("AppConfig").(*AppConfig)
	redisPool := c.MustGet("RedisPool").(*redis.Pool)
	deviceMapper := c.MustGet("DeviceMapper").(*DeviceMapper)

	var request TriggerRequest
	err := c.ShouldBind(&request)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	// Clean up request trigger type before we enqueue it
	normalTriggerType := request.NormalizeTriggerType()
	log.Printf("Normalizing Type. Was: {%s}, Is: {%s}", request.TriggerType, normalTriggerType)
	request.TriggerType = normalTriggerType

	// Clean up request trigger type before we enqueue it
	normalDeviceName, err := request.NormalizeDeviceName()
	if err != nil {
		log.Println("Could not normalize request's device name")
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	log.Printf("Normalizing Device Name. Was: {%s}, Is: {%s}\n", request.DeviceName, normalDeviceName)
	request.DeviceName = normalDeviceName

	// Map incoming device name to a supported device if possible
	supportedDeviceName := deviceMapper.MapToSupportedDevice(request.DeviceName)
	if supportedDeviceName != "" {
		log.Printf("Mapping {%s} to {%s}\n", request.DeviceName, supportedDeviceName)
		request.DeviceName = supportedDeviceName
	} else {
		log.Printf("Could not map device name: {%s}", request.DeviceName)
	}

	enqueuer := work.NewEnqueuer(appConfig.Namespace, redisPool)
	_, err = EnqueueRequest(&request, enqueuer)
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func EnqueueRequest(request *TriggerRequest, enqueuer *work.Enqueuer) (*work.ScheduledJob, error) {
	delaySecs, err := request.ConvertDelayToSeconds()
	if err != nil {
		log.Println("Could not parse request delay time")
		log.Println(err)
		return nil, err
	}

	device := request.DeviceName
	triggerType := request.TriggerType
	triggerKey := request.TriggerKey()
	log.Printf("[%s]: Scheduled {%s} to {%s} in {%d} seconds\n", triggerKey, device, triggerType, delaySecs)

	return enqueuer.EnqueueIn("delay_trigger", delaySecs, work.Q{
		"device":      device,
		"delay":       delaySecs,
		"trigger_key": triggerKey,
	})
}
