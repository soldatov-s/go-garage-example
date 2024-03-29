// Package apiv1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package apiv1

// DataResult defines model for DataResult.
type DataResult struct {
	Result *Enity `json:"result,omitempty"`
}

// Enity defines model for Enity.
type Enity struct {
	Code      *string   `db:"code" json:"code,omitempty"`
	CreatedAt *NullTime `db:"created_at" json:"created_at,omitempty"`
	DeletedAt *NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
	Id        *int64    `db:"id" json:"id,omitempty"`
	Meta      *NullMeta `db:"meta" json:"meta,omitempty"`
	UpdatedAt *NullTime `db:"updated_at" json:"updated_at,omitempty"`
}

// ErrorAnsw defines model for ErrorAnsw.
type ErrorAnsw struct {
	Error *ErrorAnswBody `json:"error,omitempty"`
}

// ErrorAnswBody defines model for ErrorAnswBody.
type ErrorAnswBody struct {
	Code       *string `json:"code,omitempty"`
	Details    *string `json:"details,omitempty"`
	StatusCode *int    `json:"statusCode,omitempty"`
}

// ResultAnsw defines model for ResultAnsw.
type ResultAnsw struct {
	Result *string `json:"result,omitempty"`
}

// PostHandlerJSONBody defines parameters for PostHandler.
type PostHandlerJSONBody Enity

// DeleteHandlerParams defines parameters for DeleteHandler.
type DeleteHandlerParams struct {
	// Hard delete data, if equal true, delete hard
	Hard *bool `json:"hard,omitempty"`
}

// PostHandlerJSONRequestBody defines body for PostHandler for application/json ContentType.
type PostHandlerJSONRequestBody PostHandlerJSONBody

