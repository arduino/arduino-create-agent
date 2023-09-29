// Code generated by goa v3.13.1, DO NOT EDIT.
//
// indexes HTTP server types
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package server

import (
	indexes "github.com/arduino/arduino-create-agent/gen/indexes"
	indexesviews "github.com/arduino/arduino-create-agent/gen/indexes/views"
	goa "goa.design/goa/v3/pkg"
)

// AddRequestBody is the type of the "indexes" service "add" endpoint HTTP
// request body.
type AddRequestBody struct {
	// The url of the index file
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
}

// RemoveRequestBody is the type of the "indexes" service "remove" endpoint
// HTTP request body.
type RemoveRequestBody struct {
	// The url of the index file
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
}

// AddResponseBody is the type of the "indexes" service "add" endpoint HTTP
// response body.
type AddResponseBody struct {
	// The status of the operation
	Status string `form:"status" json:"status" xml:"status"`
}

// RemoveResponseBody is the type of the "indexes" service "remove" endpoint
// HTTP response body.
type RemoveResponseBody struct {
	// The status of the operation
	Status string `form:"status" json:"status" xml:"status"`
}

// ListInvalidURLResponseBody is the type of the "indexes" service "list"
// endpoint HTTP response body for the "invalid_url" error.
type ListInvalidURLResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// AddInvalidURLResponseBody is the type of the "indexes" service "add"
// endpoint HTTP response body for the "invalid_url" error.
type AddInvalidURLResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// RemoveInvalidURLResponseBody is the type of the "indexes" service "remove"
// endpoint HTTP response body for the "invalid_url" error.
type RemoveInvalidURLResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// NewAddResponseBody builds the HTTP response body from the result of the
// "add" endpoint of the "indexes" service.
func NewAddResponseBody(res *indexesviews.OperationView) *AddResponseBody {
	body := &AddResponseBody{
		Status: *res.Status,
	}
	return body
}

// NewRemoveResponseBody builds the HTTP response body from the result of the
// "remove" endpoint of the "indexes" service.
func NewRemoveResponseBody(res *indexesviews.OperationView) *RemoveResponseBody {
	body := &RemoveResponseBody{
		Status: *res.Status,
	}
	return body
}

// NewListInvalidURLResponseBody builds the HTTP response body from the result
// of the "list" endpoint of the "indexes" service.
func NewListInvalidURLResponseBody(res *goa.ServiceError) *ListInvalidURLResponseBody {
	body := &ListInvalidURLResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewAddInvalidURLResponseBody builds the HTTP response body from the result
// of the "add" endpoint of the "indexes" service.
func NewAddInvalidURLResponseBody(res *goa.ServiceError) *AddInvalidURLResponseBody {
	body := &AddInvalidURLResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewRemoveInvalidURLResponseBody builds the HTTP response body from the
// result of the "remove" endpoint of the "indexes" service.
func NewRemoveInvalidURLResponseBody(res *goa.ServiceError) *RemoveInvalidURLResponseBody {
	body := &RemoveInvalidURLResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewAddIndexPayload builds a indexes service add endpoint payload.
func NewAddIndexPayload(body *AddRequestBody) *indexes.IndexPayload {
	v := &indexes.IndexPayload{
		URL: *body.URL,
	}

	return v
}

// NewRemoveIndexPayload builds a indexes service remove endpoint payload.
func NewRemoveIndexPayload(body *RemoveRequestBody) *indexes.IndexPayload {
	v := &indexes.IndexPayload{
		URL: *body.URL,
	}

	return v
}

// ValidateAddRequestBody runs the validations defined on AddRequestBody
func ValidateAddRequestBody(body *AddRequestBody) (err error) {
	if body.URL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("url", "body"))
	}
	return
}

// ValidateRemoveRequestBody runs the validations defined on RemoveRequestBody
func ValidateRemoveRequestBody(body *RemoveRequestBody) (err error) {
	if body.URL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("url", "body"))
	}
	return
}
