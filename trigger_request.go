package triggo

import "time"

type TriggerRequest struct {
	Device         string    `json:"device"`
	Delay          float64   `json:"minutes,float64"`
	CreatedTimeStr string    `json:"created_time_str"`
	CreatedTime    time.Time `json:"created_time, time_format:"unix"`
}
