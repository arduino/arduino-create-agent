package v2

import (
	"context"
	"net/http"

	"github.com/Sirupsen/logrus"
	toolssvr "github.com/arduino/arduino-create-agent/gen/http/tools/server"
	toolssvc "github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/tools"
	goahttp "goa.design/goa/http"
	"goa.design/goa/http/middleware"
)

func Server() http.Handler {
	mux := goahttp.NewMuxer()

	// Instantiate logger
	logger := logrus.New()
	logAdapter := LogAdapter{Logger: logger}

	// Mount tools
	toolsSvc := tools.Tools{}
	toolsEndpoints := toolssvc.NewEndpoints(&toolsSvc)

	toolsServer := toolssvr.New(toolsEndpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, errorHandler(logger))
	toolssvr.Mount(mux, toolsServer)

	// Mount middlewares
	handler := middleware.Log(logAdapter)(mux)
	handler = middleware.RequestID()(handler)

	return handler
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *logrus.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
