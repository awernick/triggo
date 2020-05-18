package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

type Context struct {
	IFTTTAPIKey string
	IFTTTAPIURL string
}

func RunAsWorkerNode(appConfig AppConfig, redisPool *redis.Pool) {
	var err error

	// Validate IFTTT API Key
	err = appConfig.ValidateIFTTTAPIKey()
	if err != nil {
		log.Fatal(err)
	}

	// Validate IFTTT API URL
	err = appConfig.ValidateIFTTTAPIURL()
	if err != nil {
		log.Fatal(err)
	}

	// Inject API Key and URL to Job Context
	ctx := Context{
		IFTTTAPIKey: appConfig.IFTTTAPIKey,
		IFTTTAPIURL: appConfig.IFTTTAPIURL,
	}

	// Start background worker pool
	pool := work.NewWorkerPool(ctx, 10, appConfig.Namespace, redisPool)
	pool.Job("delay_trigger", ctx.ProcessTriggerRequest)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

func (c *Context) ProcessTriggerRequest(job *work.Job) error {
	device := job.ArgString("device")
	delay := job.ArgInt64("delay")
	triggerKey := job.ArgString("trigger_key")

	if err := job.ArgError(); err != nil {
		log.Println(err)
		return err
	}

	// I need to map a device to callback
	log.Println(fmt.Sprintf("Device: %s", device))
	log.Println(fmt.Sprintf("Delay: %d", delay))
	log.Println(fmt.Sprintf("Trigger Key: %s", triggerKey))
	log.Println((*c).IFTTTAPIURL)
	log.Println((*c).IFTTTAPIKey)
	requestURL, _ := url.Parse((*c).IFTTTAPIURL)
	requestURL.Path = path.Join(fmt.Sprintf(IFTTTTriggerURLPath, triggerKey, (*c).IFTTTAPIKey))

	log.Printf("POSTing to %s", requestURL.String())
	client := &http.Client{}
	req, err := http.NewRequest("POST", requestURL.String(), nil)
	if err != nil {
		log.Print(err)
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return err
	}

	bodyString := string(body)
	log.Println(bodyString)

	if res.StatusCode == 200 {
		log.Print("Everything worked!")
	}

	return nil
}
