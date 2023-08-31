// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	indexessvr "github.com/arduino/arduino-create-agent/gen/http/indexes/server"
	toolssvr "github.com/arduino/arduino-create-agent/gen/http/tools/server"
	indexessvc "github.com/arduino/arduino-create-agent/gen/indexes"
	toolssvc "github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/sirupsen/logrus"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/http/middleware"
	goamiddleware "goa.design/goa/v3/middleware"
)

// Server is the actual server
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
		goahttp.ResponseEncoder, errorHandler(logger), nil)
	indexessvr.Mount(mux, indexesServer)

	// Mount tools
	toolsSvc := pkgs.Tools{
		Folder:  home,
		Indexes: &indexesSvc,
	}
	toolsEndpoints := toolssvc.NewEndpoints(&toolsSvc)
	toolsServer := toolssvr.New(toolsEndpoints, mux, CustomRequestDecoder, goahttp.ResponseEncoder, errorHandler(logger), nil)
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
		id := ctx.Value(goamiddleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}

// CustomRequestDecoder overrides the RequestDecoder provided by goahttp package
// It returns always a json.NewDecoder for legacy reasons:
// The web editor sends always request to the agent setting "Content-Type: text/plain"
// even when it should set "Content-Type: application/json". This breaks the parsing with:
// "can't decode text/plain to *server.InstallRequestBody" error message.
// This was working before the bump to goa v3 only because a "text/plain" decoder was not implemented
// and it was fallbacking to the json decoder. (https://github.com/goadesign/goa/pull/2310)
func CustomRequestDecoder(r *http.Request) goahttp.Decoder {
	return json.NewDecoder(r.Body)
}
