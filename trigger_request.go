package main

import "time"

type TriggerRequest struct {
	Device         string    `json:"device"`
	DelayInMins    string    `json:"delay_mins,omitempty"`
	Delay          int64     `json:"delay,int64"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time, time_format:"unix"`
}
