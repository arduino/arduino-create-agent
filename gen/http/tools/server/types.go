// Code generated by goa v3.12.4, DO NOT EDIT.
//
// tools HTTP server types
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package server

import (
	tools "github.com/arduino/arduino-create-agent/gen/tools"
	toolsviews "github.com/arduino/arduino-create-agent/gen/tools/views"
	goa "goa.design/goa/v3/pkg"
)

// InstallRequestBody is the type of the "tools" service "install" endpoint
// HTTP request body.
type InstallRequestBody struct {
	// The name of the tool
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the tool
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
	// The packager of the tool
	Packager *string `form:"packager,omitempty" json:"packager,omitempty" xml:"packager,omitempty"`
}

// ToolResponseCollection is the type of the "tools" service "available"
// endpoint HTTP response body.
type ToolResponseCollection []*ToolResponse

// InstallResponseBody is the type of the "tools" service "install" endpoint
// HTTP response body.
type InstallResponseBody struct {
	// The status of the operation
	Status string `form:"status" json:"status" xml:"status"`
}

// RemoveResponseBody is the type of the "tools" service "remove" endpoint HTTP
// response body.
type RemoveResponseBody struct {
	// The status of the operation
	Status string `form:"status" json:"status" xml:"status"`
}

// ToolResponse is used to define fields on response body types.
type ToolResponse struct {
	// The name of the tool
	Name string `form:"name" json:"name" xml:"name"`
	// The version of the tool
	Version string `form:"version" json:"version" xml:"version"`
	// The packager of the tool
	Packager string `form:"packager" json:"packager" xml:"packager"`
}

// NewToolResponseCollection builds the HTTP response body from the result of
// the "available" endpoint of the "tools" service.
func NewToolResponseCollection(res toolsviews.ToolCollectionView) ToolResponseCollection {
	body := make([]*ToolResponse, len(res))
	for i, val := range res {
		body[i] = marshalToolsviewsToolViewToToolResponse(val)
	}
	return body
}

// NewInstallResponseBody builds the HTTP response body from the result of the
// "install" endpoint of the "tools" service.
func NewInstallResponseBody(res *toolsviews.OperationView) *InstallResponseBody {
	body := &InstallResponseBody{
		Status: *res.Status,
	}
	return body
}

// NewRemoveResponseBody builds the HTTP response body from the result of the
// "remove" endpoint of the "tools" service.
func NewRemoveResponseBody(res *toolsviews.OperationView) *RemoveResponseBody {
	body := &RemoveResponseBody{
		Status: *res.Status,
	}
	return body
}

// NewInstallToolPayload builds a tools service install endpoint payload.
func NewInstallToolPayload(body *InstallRequestBody) *tools.ToolPayload {
	v := &tools.ToolPayload{
		Name:     *body.Name,
		Version:  *body.Version,
		Packager: *body.Packager,
	}

	return v
}

// NewRemoveToolPayload builds a tools service remove endpoint payload.
func NewRemoveToolPayload(packager string, name string, version string) *tools.ToolPayload {
	v := &tools.ToolPayload{}
	v.Packager = packager
	v.Name = name
	v.Version = version

	return v
}

// ValidateInstallRequestBody runs the validations defined on InstallRequestBody
func ValidateInstallRequestBody(body *InstallRequestBody) (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.Version == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("version", "body"))
	}
	if body.Packager == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("packager", "body"))
	}
	return
}
