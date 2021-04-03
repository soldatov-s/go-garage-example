package test

import (
	"encoding/json"
	"time"

	"github.com/soldatov-s/go-garage/models"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/types"
	"github.com/soldatov-s/go-garage/x/sql"
)

// Private type, used for configurate logger
type empty struct{}

type Enity struct {
	ID   int            `json:"id" db:"id"`
	Code string         `json:"code" db:"code"`
	Meta types.NullMeta `json:"meta" db:"meta"`
	models.Timestamp
}

func (e *Enity) SQLParamsRequest(h *sql.Helper) []string {
	return h.RequestParamsWithout(e, "id")
}

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

type DataResult httpsrv.ResultAnsw
