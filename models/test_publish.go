package models

import "time"

type TestPublish struct {
	Code   string    `json:"code"`
	SendAt time.Time `json:"send_at"`
}
