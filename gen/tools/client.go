// Code generated by goa v3.13.2, DO NOT EDIT.
//
// tools client
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package tools

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Client is the "tools" service client.
type Client struct {
	AvailableEndpoint     goa.Endpoint
	InstalledheadEndpoint goa.Endpoint
	InstalledEndpoint     goa.Endpoint
	InstallEndpoint       goa.Endpoint
	RemoveEndpoint        goa.Endpoint
}

// NewClient initializes a "tools" service client given the endpoints.
func NewClient(available, installedhead, installed, install, remove goa.Endpoint) *Client {
	return &Client{
		AvailableEndpoint:     available,
		InstalledheadEndpoint: installedhead,
		InstalledEndpoint:     installed,
		InstallEndpoint:       install,
		RemoveEndpoint:        remove,
	}
}

// Available calls the "available" endpoint of the "tools" service.
func (c *Client) Available(ctx context.Context) (res ToolCollection, err error) {
	var ires any
	ires, err = c.AvailableEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(ToolCollection), nil
}

// Installedhead calls the "installedhead" endpoint of the "tools" service.
func (c *Client) Installedhead(ctx context.Context) (err error) {
	_, err = c.InstalledheadEndpoint(ctx, nil)
	return
}

// Installed calls the "installed" endpoint of the "tools" service.
func (c *Client) Installed(ctx context.Context) (res ToolCollection, err error) {
	var ires any
	ires, err = c.InstalledEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(ToolCollection), nil
}

// Install calls the "install" endpoint of the "tools" service.
// Install may return the following errors:
//   - "not_found" (type *goa.ServiceError): tool not found
//   - error: internal error
func (c *Client) Install(ctx context.Context, p *ToolPayload) (res *Operation, err error) {
	var ires any
	ires, err = c.InstallEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*Operation), nil
}

// Remove calls the "remove" endpoint of the "tools" service.
func (c *Client) Remove(ctx context.Context, p *ToolPayload) (res *Operation, err error) {
	var ires any
	ires, err = c.RemoveEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*Operation), nil
}
