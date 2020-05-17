package main

import (
	"regexp"
	"strconv"
	"time"

	"github.com/gobuffalo/flect"
)

const IFTTTTriggerURLPath = "/trigger/%s/with/key/%s"

const DefiniteArticleRegex = `^(a|an|and|the)(\s+)`

type TriggerRequest struct {
	TriggerType    string    `json:"trigger_type" binding:"required"`
	DeviceName     string    `json:"device" binding:"required"`
	Delay          string    `json:"delay_mins" binding:"required"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time" time_format:"unix"`
	SecretKey      string    `json:"secret_key"`
}

func (tr *TriggerRequest) NormalizedDeviceName() string {
	rx := regexp.MustCompile(DefiniteArticleRegex)
	deviceName := rx.ReplaceAllString(tr.DeviceName, "")
	deviceName = flect.Singularize(deviceName)
	return flect.Underscore(deviceName)
}

func (tr *TriggerRequest) NormalizedTriggerType() string {
	return flect.Underscore(tr.TriggerType)
}

func (tr *TriggerRequest) ConvertDelayToSeconds() (int64, error) {
	delay, err := strconv.ParseInt(tr.Delay, 10, 64)
	if err != nil {
		return 0, err
	} else {
		return delay * 60, nil
	}
}
