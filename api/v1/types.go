package apiv1

import garageTypes "github.com/soldatov-s/go-garage/types"

type NullMeta struct {
	garageTypes.NullMeta
}

func NewNullMeta() *NullMeta {
	v := &NullMeta{}
	v.Valid = true
	return v
}

type NullTime struct {
	garageTypes.NullTime
}

func NewNullTime() *NullTime {
	v := &NullTime{}
	v.Timestamp()
	return v
}

type NullString struct {
	garageTypes.NullString
}

func NewNullString() *NullString {
	v := &NullString{}
	v.Valid = true
	return v
}
