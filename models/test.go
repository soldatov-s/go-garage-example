package models

import (
	"encoding/json"

	"github.com/soldatov-s/go-garage/models"
	"github.com/soldatov-s/go-garage/types"
)

type Test struct {
	ID   int            `json:"id" db:"id"`
	Code string         `json:"code" db:"code"`
	Meta types.NullMeta `json:"meta" db:"meta"`
	models.Timestamp
}

func (t *Test) SQLParamsRequest() []string {
	return []string{
		"code",
		"meta",
		"created_at",
		"updated_at",
		"deleted_at",
	}
}

func (t *Test) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Test) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}
