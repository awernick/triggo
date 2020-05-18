package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file or missing. Skipping...")
	}

	// Parse CLI flags
	redisURL := flag.String("redis-url", os.Getenv("REDIS_URL"), "url of Redis instance [required]")
	isWorker := flag.Bool("worker", false, "run as background worker node")
	secretKey := flag.String("secret-key", os.Getenv("SECRET_KEY"), "key used to authenticate requests")
	iftttAPIKey := flag.String("ifttt-api-key", os.Getenv("IFTTT_API_KEY"), "API Key for IFTTT [required for worker]")
	iftttAPIURL := flag.String("ifttt-api-url", os.Getenv("IFTTT_API_URL"), "API URL for IFTTT [required for worker]")
	httpPort := flag.String("port", os.Getenv("PORT"), "port used to listen for incoming http requests [required for server]")
	namespace := flag.String("namespace", ProgName(), "namespace used for redis work queues. Defaults to program name (arg[0]).")
	flag.Parse()

	appConfig := AppConfig{
		IFTTTAPIKey: *iftttAPIKey,
		IFTTTAPIURL: *iftttAPIURL,
		SecretKey:   *secretKey,
		HTTPPort:    *httpPort,
		Namespace:   *namespace,
	}

	// Instantiate a Redis pool with 5 connection
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(*redisURL)
		},
	}

	if *isWorker {
		log.Println("Running as Worker Node")
		RunAsWorkerNode(appConfig, redisPool)
	} else {
		log.Println("Running as Server Node")
		RunAsServerNode(appConfig, redisPool)
	}
}

func ProgName() string {
	return filepath.Base(os.Args[0])
}
