package apiv1

import (
	"encoding/json"
	"time"
)

func (e *Enity) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Enity) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

type Consume struct {
	Code   string `json:"code"`
	SendAt int    `json:"send_at"`
}

type Publish struct {
	Code   string    `json:"code"`
	SendAt time.Time `json:"send_at"`
}
