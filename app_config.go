package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
)

type AppConfig struct {
	IFTTTAPIKey string
	IFTTTAPIURL string
	SecretKey   string
	HTTPPort    string
	Namespace   string
}

func (ac *AppConfig) ValidateIFTTTAPIKey() error {
	if len(ac.IFTTTAPIKey) == 0 {
		return errors.New("please specify an IFTTT API Key")
	}

	return nil
}

func (ac *AppConfig) ValidateIFTTTAPIURL() error {
	if len(ac.IFTTTAPIURL) == 0 {
		return errors.New("please specify an IFTTT API URL")
	}

	_, err := url.Parse(ac.IFTTTAPIURL)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("invalid IFTTT API URL: %s", ac.IFTTTAPIURL)
	}

	return nil
}

func (ac *AppConfig) ValidateHTTPPort() error {
	if len(ac.HTTPPort) == 0 {
		return errors.New("please specify HTTP Port")
	}

	return nil
}
