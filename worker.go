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
)

type Context struct{}

func RunAsWorkerNode() {

	if len((*iftttConfig).APIKey) == 0 {
		log.Fatal("Please specify an IFTTT API Key.")
	}

	if len((*iftttConfig).APIURL) == 0 {
		log.Fatal("Please specify an IFTTT API URL.")
	}

	_, err := url.Parse((*iftttConfig).APIURL)
	if err != nil {
		log.Println("Invalid IFTTT API URL")
		log.Fatal(err)
	}

	pool := work.NewWorkerPool(Context{}, 10, QueueNamespace, redisPool)
	pool.Job("delay_trigger", (*Context).ProcessTriggerRequest)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
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

	requestURL, _ := url.Parse((*iftttConfig).APIURL)
	requestURL.Path = path.Join(requestURL.Path, triggerKey, "with", "key", iftttConfig.APIKey)

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
