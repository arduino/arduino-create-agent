package v2

import (
	"net/http"

	toolssvr "github.com/arduino/arduino-create-agent/gen/http/tools/server"
	toolssvc "github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/tools"
	goahttp "goa.design/goa/http"
)

func Server() http.Handler {
	mux := goahttp.NewMuxer()

	// Mount tools
	toolsSvc := tools.Tools{}
	toolsEndpoints := toolssvc.NewEndpoints(&toolsSvc)

	toolsServer := toolssvr.New(toolsEndpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
	toolssvr.Mount(mux, toolsServer)

	return mux
}
