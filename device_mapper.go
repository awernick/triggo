package main

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type DeviceMapper struct {
	mappings map[string]string
}

func (dm *DeviceMapper) MapToSupportedDevice(key string) string {
	return dm.mappings[key]
}

func (dm *DeviceMapper) LoadMappings() error {
	data, err := ioutil.ReadFile("mappings.yaml")
	if err != nil {
		log.Println(err)
		return errors.New("could not load device mappings.yaml")
	}

	err = yaml.Unmarshal([]byte(data), &dm.mappings)
	if err != nil {
		log.Println(err)
		return errors.New("could not load device mappings in to memory")
	}

	return nil
}
