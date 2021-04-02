package test

import "time"

type Publish struct {
	Code   string    `json:"code"`
	SendAt time.Time `json:"send_at"`
}
