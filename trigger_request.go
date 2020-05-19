package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/flect"
	"gopkg.in/jdkato/prose.v2"
)

const IFTTTTriggerURLPath = "/trigger/%s/with/key/%s"

type TriggerRequest struct {
	TriggerType    string    `json:"trigger_type" binding:"required"`
	DeviceName     string    `json:"device" binding:"required"`
	Delay          string    `json:"delay_mins" binding:"required"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time" time_format:"unix"`
	SecretKey      string    `json:"secret_key"`
}

func (tr TriggerRequest) extractDeviceNameNouns() ([]string, error) {
	var nouns []string

	doc, err := prose.NewDocument(tr.DeviceName)
	if err != nil {
		return nouns, err
	}

	for _, tok := range doc.Tokens() {
		if strings.Contains(tok.Tag, "NN") {
			nouns = append(nouns, tok.Text)
		}
	}

	return nouns, nil
}

func (tr TriggerRequest) NormalizeDeviceName() (string, error) {
	nouns, err := tr.extractDeviceNameNouns()
	if err != nil {
		return "", err
	}
	cleanDeviceName := strings.Join(nouns, " ")
	singularDeviceName := flect.Singularize(cleanDeviceName)
	normalizedDeviceName := flect.Underscore(singularDeviceName)
	return normalizedDeviceName, nil
}

func (tr *TriggerRequest) NormalizeTriggerType() string {
	return flect.Underscore(tr.TriggerType)
}

func (tr TriggerRequest) TriggerKey() string {
	return strings.Join([]string{tr.TriggerType, tr.DeviceName}, "_")
}

func (tr *TriggerRequest) ConvertDelayToSeconds() (int64, error) {
	delay, err := strconv.ParseInt(tr.Delay, 10, 64)
	if err != nil {
		return 0, err
	} else {
		return delay * 60, nil
	}
}
