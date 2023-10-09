// Code generated by goa v3.13.2, DO NOT EDIT.
//
// tools HTTP server encoders and decoders
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package server

import (
	"context"
	"io"
	"net/http"

	toolsviews "github.com/arduino/arduino-create-agent/gen/tools/views"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// EncodeAvailableResponse returns an encoder for responses returned by the
// tools available endpoint.
func EncodeAvailableResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res := v.(toolsviews.ToolCollection)
		enc := encoder(ctx, w)
		body := NewToolResponseCollection(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// EncodeInstalledheadResponse returns an encoder for responses returned by the
// tools installedhead endpoint.
func EncodeInstalledheadResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

// EncodeInstalledResponse returns an encoder for responses returned by the
// tools installed endpoint.
func EncodeInstalledResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res := v.(toolsviews.ToolCollection)
		enc := encoder(ctx, w)
		body := NewToolResponseCollection(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// EncodeInstallResponse returns an encoder for responses returned by the tools
// install endpoint.
func EncodeInstallResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res := v.(*toolsviews.Operation)
		enc := encoder(ctx, w)
		body := NewInstallResponseBody(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeInstallRequest returns a decoder for requests sent to the tools
// install endpoint.
func DecodeInstallRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (any, error) {
	return func(r *http.Request) (any, error) {
		var (
			body InstallRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = ValidateInstallRequestBody(&body)
		if err != nil {
			return nil, err
		}
		payload := NewInstallToolPayload(&body)

		return payload, nil
	}
}

// EncodeRemoveResponse returns an encoder for responses returned by the tools
// remove endpoint.
func EncodeRemoveResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res := v.(*toolsviews.Operation)
		enc := encoder(ctx, w)
		body := NewRemoveResponseBody(res.Projected)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeRemoveRequest returns a decoder for requests sent to the tools remove
// endpoint.
func DecodeRemoveRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (any, error) {
	return func(r *http.Request) (any, error) {
		var (
			body RemoveRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}

		var (
			packager string
			name     string
			version  string

			params = mux.Vars(r)
		)
		packager = params["packager"]
		name = params["name"]
		version = params["version"]
		payload := NewRemoveToolPayload(&body, packager, name, version)

		return payload, nil
	}
}

// marshalToolsviewsToolViewToToolResponse builds a value of type *ToolResponse
// from a value of type *toolsviews.ToolView.
func marshalToolsviewsToolViewToToolResponse(v *toolsviews.ToolView) *ToolResponse {
	res := &ToolResponse{
		Name:     *v.Name,
		Version:  *v.Version,
		Packager: *v.Packager,
	}

	return res
}
