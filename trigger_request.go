package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/flect"
)

type TriggerRequest struct {
	TriggerType    string    `json:"trigger_type" binding:"required"`
	DeviceName     string    `json:"device" binding:"required"`
	DelayInMins    string    `json:"delay_mins" binding:"required"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time" time_format:"unix"`
}

const DefiniteArticleRegex = `(a|an|and|the)(\s*)`

func (tr *TriggerRequest) NormalizedDeviceName() string {
	rx := regexp.MustCompile(DefiniteArticleRegex)
	return rx.ReplaceAllString(tr.DeviceName, "")
}

func (tr *TriggerRequest) TriggerKey() string {
	strs := []string{flect.Underscore(tr.TriggerType), flect.Underscore(tr.NormalizedDeviceName())}
	return strings.Join(strs, "_")
}

func (tr *TriggerRequest) Delay() (int64, error) {
	delay, err := strconv.ParseInt(tr.DelayInMins, 10, 64)
	if err != nil {
		return 0, err
	} else {
		return delay * 60, nil
	}
}
