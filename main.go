package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

type IFTTTConfig struct {
	APIKey string
	APIURL string
}

var iftttConfig *IFTTTConfig
var redisPool *redis.Pool

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file or missing. Skipping...")
	}

	// Parse CLI flags
	redisURL := flag.String("redis-url", os.Getenv("REDIS_URL"), "url of Redis instance [required]")
	isWorker := flag.Bool("worker", false, "run as background worker node")
	iftttAPIKey := flag.String("ifttt-api-key", os.Getenv("IFTTT_API_KEY"), "API Key for IFTTT [required for worker]")
	iftttAPIURL := flag.String("ifttt-api-url", os.Getenv("IFTTT_API_URL"), "API URL for IFTTT [required for worker]")
	flag.Parse()

	iftttConfig = &IFTTTConfig{
		APIKey: *iftttAPIKey,
		APIURL: *iftttAPIURL,
	}

	// Instantiate a Redis pool with 5 connection
	redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(*redisURL)
		},
	}

	if *isWorker {
		log.Println("Running as Worker Node")
		RunAsWorkerNode()
	} else {
		log.Println("Running as Server Node")
		RunAsServerNode()
	}
}

func Namespace() string {
	return filepath.Base(os.Args[0])
}
