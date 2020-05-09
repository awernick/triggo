package triggo

import (
	"flag"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

var redisPool *redis.Pool

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Parse CLI flags
	redisURL := flag.String("redis-url", os.Getenv("REDIS_URL"), "url of Redis instance")
	isWorker := flag.Bool("worker", false, "run as Sidekiq worker node")
	flag.Parse()

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
		RunAsWorkerNode(redisPool)
	} else {
		log.Println("Running as Server Node")
		RunAsServerNode(redisPool)
	}
}
