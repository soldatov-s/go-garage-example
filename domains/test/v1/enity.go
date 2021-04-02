package test

import (
	"context"
	"encoding/json"

	"github.com/soldatov-s/go-garage/models"
	"github.com/soldatov-s/go-garage/types"
	"github.com/soldatov-s/go-garage/x/sql"
)

type Enity struct {
	ID   int            `json:"id" db:"id"`
	Code string         `json:"code" db:"code"`
	Meta types.NullMeta `json:"meta" db:"meta"`
	models.Timestamp
}

func (t *Enity) SQLParamsRequest(ctx context.Context) []string {
	return sql.Get(ctx).RequestParamsWithout(t, "id")
}

func (t *Enity) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Enity) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}
