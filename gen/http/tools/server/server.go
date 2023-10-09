// Code generated by goa v3.13.2, DO NOT EDIT.
//
// tools HTTP server
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package server

import (
	"context"
	"net/http"

	tools "github.com/arduino/arduino-create-agent/gen/tools"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// Server lists the tools service endpoint HTTP handlers.
type Server struct {
	Mounts        []*MountPoint
	Available     http.Handler
	Installedhead http.Handler
	Installed     http.Handler
	Install       http.Handler
	Remove        http.Handler
}

// MountPoint holds information about the mounted endpoints.
type MountPoint struct {
	// Method is the name of the service method served by the mounted HTTP handler.
	Method string
	// Verb is the HTTP method used to match requests to the mounted handler.
	Verb string
	// Pattern is the HTTP request path pattern used to match requests to the
	// mounted handler.
	Pattern string
}

// New instantiates HTTP handlers for all the tools service endpoints using the
// provided encoder and decoder. The handlers are mounted on the given mux
// using the HTTP verb and path defined in the design. errhandler is called
// whenever a response fails to be encoded. formatter is used to format errors
// returned by the service methods prior to encoding. Both errhandler and
// formatter are optional and can be nil.
func New(
	e *tools.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"Available", "GET", "/v2/pkgs/tools/available"},
			{"Installedhead", "HEAD", "/v2/pkgs/tools/installed"},
			{"Installed", "GET", "/v2/pkgs/tools/installed"},
			{"Install", "POST", "/v2/pkgs/tools/installed"},
			{"Remove", "DELETE", "/v2/pkgs/tools/installed/{packager}/{name}/{version}"},
		},
		Available:     NewAvailableHandler(e.Available, mux, decoder, encoder, errhandler, formatter),
		Installedhead: NewInstalledheadHandler(e.Installedhead, mux, decoder, encoder, errhandler, formatter),
		Installed:     NewInstalledHandler(e.Installed, mux, decoder, encoder, errhandler, formatter),
		Install:       NewInstallHandler(e.Install, mux, decoder, encoder, errhandler, formatter),
		Remove:        NewRemoveHandler(e.Remove, mux, decoder, encoder, errhandler, formatter),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "tools" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.Available = m(s.Available)
	s.Installedhead = m(s.Installedhead)
	s.Installed = m(s.Installed)
	s.Install = m(s.Install)
	s.Remove = m(s.Remove)
}

// MethodNames returns the methods served.
func (s *Server) MethodNames() []string { return tools.MethodNames[:] }

// Mount configures the mux to serve the tools endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountAvailableHandler(mux, h.Available)
	MountInstalledheadHandler(mux, h.Installedhead)
	MountInstalledHandler(mux, h.Installed)
	MountInstallHandler(mux, h.Install)
	MountRemoveHandler(mux, h.Remove)
}

// Mount configures the mux to serve the tools endpoints.
func (s *Server) Mount(mux goahttp.Muxer) {
	Mount(mux, s)
}

// MountAvailableHandler configures the mux to serve the "tools" service
// "available" endpoint.
func MountAvailableHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/v2/pkgs/tools/available", f)
}

// NewAvailableHandler creates a HTTP handler which loads the HTTP request and
// calls the "tools" service "available" endpoint.
func NewAvailableHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		encodeResponse = EncodeAvailableResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "available")
		ctx = context.WithValue(ctx, goa.ServiceKey, "tools")
		var err error
		res, err := endpoint(ctx, nil)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountInstalledheadHandler configures the mux to serve the "tools" service
// "installedhead" endpoint.
func MountInstalledheadHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("HEAD", "/v2/pkgs/tools/installed", f)
}

// NewInstalledheadHandler creates a HTTP handler which loads the HTTP request
// and calls the "tools" service "installedhead" endpoint.
func NewInstalledheadHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		encodeResponse = EncodeInstalledheadResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "installedhead")
		ctx = context.WithValue(ctx, goa.ServiceKey, "tools")
		var err error
		res, err := endpoint(ctx, nil)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountInstalledHandler configures the mux to serve the "tools" service
// "installed" endpoint.
func MountInstalledHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/v2/pkgs/tools/installed", f)
}

// NewInstalledHandler creates a HTTP handler which loads the HTTP request and
// calls the "tools" service "installed" endpoint.
func NewInstalledHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		encodeResponse = EncodeInstalledResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "installed")
		ctx = context.WithValue(ctx, goa.ServiceKey, "tools")
		var err error
		res, err := endpoint(ctx, nil)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountInstallHandler configures the mux to serve the "tools" service
// "install" endpoint.
func MountInstallHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/v2/pkgs/tools/installed", f)
}

// NewInstallHandler creates a HTTP handler which loads the HTTP request and
// calls the "tools" service "install" endpoint.
func NewInstallHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeInstallRequest(mux, decoder)
		encodeResponse = EncodeInstallResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "install")
		ctx = context.WithValue(ctx, goa.ServiceKey, "tools")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		res, err := endpoint(ctx, payload)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountRemoveHandler configures the mux to serve the "tools" service "remove"
// endpoint.
func MountRemoveHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("DELETE", "/v2/pkgs/tools/installed/{packager}/{name}/{version}", f)
}

// NewRemoveHandler creates a HTTP handler which loads the HTTP request and
// calls the "tools" service "remove" endpoint.
func NewRemoveHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeRemoveRequest(mux, decoder)
		encodeResponse = EncodeRemoveResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "remove")
		ctx = context.WithValue(ctx, goa.ServiceKey, "tools")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		res, err := endpoint(ctx, payload)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}
