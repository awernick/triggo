package main

import (
	"strings"
	"time"

	"github.com/gobuffalo/flect"
)

type TriggerRequest struct {
	TriggerType    string    `json:"trigger_type" binding:"required"`
	Device         string    `json:"device" binding:"required"`
	DelayInMins    int64     `json:"delay_mins" binding:"required"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time" time_format:"unix"`
}

func (tr *TriggerRequest) TriggerKey() string {
	strs := []string{flect.Underscore(tr.TriggerType), flect.Underscore(tr.Device)}
	return strings.Join(strs, "_")
}

func (tr *TriggerRequest) Delay() int64 {
	return tr.DelayInMins * 60
}
