package v2

import (
	"context"
	"net/http"
	"path/filepath"

	docssvr "github.com/arduino/arduino-create-agent/gen/http/docs/server"
	indexessvr "github.com/arduino/arduino-create-agent/gen/http/indexes/server"
	toolssvr "github.com/arduino/arduino-create-agent/gen/http/tools/server"
	indexessvc "github.com/arduino/arduino-create-agent/gen/indexes"
	toolssvc "github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/sirupsen/logrus"
	goahttp "goa.design/goa/http"
	"goa.design/goa/http/middleware"
)

func Server(home string) http.Handler {
	mux := goahttp.NewMuxer()

	// Instantiate logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logAdapter := LogAdapter{Logger: logger}

	// Mount indexes
	indexesSvc := pkgs.Indexes{
		Log:    logger,
		Folder: filepath.Join(home, "indexes"),
	}
	indexesEndpoints := indexessvc.NewEndpoints(&indexesSvc)
	indexesServer := indexessvr.New(indexesEndpoints, mux, goahttp.RequestDecoder,
		goahttp.ResponseEncoder, errorHandler(logger))
	indexessvr.Mount(mux, indexesServer)

	// Mount tools
	toolsSvc := pkgs.Tools{
		Folder:  home,
		Indexes: &indexesSvc,
	}
	toolsEndpoints := toolssvc.NewEndpoints(&toolsSvc)
	toolsServer := toolssvr.New(toolsEndpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, errorHandler(logger))
	toolssvr.Mount(mux, toolsServer)

	// Mount docs
	docssvr.New(nil, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, errorHandler(logger))
	docssvr.Mount(mux)

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
