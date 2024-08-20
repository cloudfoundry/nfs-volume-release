package driveradminhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"github.com/tedsuo/rata"
)

func NewHandler(logger lager.Logger, client driveradmin.DriverAdmin) (http.Handler, error) {
	logger = logger.Session("server")
	logger.Info("start")
	defer logger.Info("end")

	var handlers = rata.Handlers{
		driveradmin.EvacuateRoute: newEvacuateHandler(logger, client),
		driveradmin.PingRoute:     newPingHandler(logger, client),
	}

	return rata.NewRouter(driveradmin.Routes, handlers)
}

func newEvacuateHandler(logger lager.Logger, client driveradmin.DriverAdmin) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("handle-evacuate")
		logger.Info("start")
		defer logger.Info("end")

		env := driverhttp.EnvWithMonitor(logger, req.Context(), w)

		response := client.Evacuate(env)
		if response.Err != "" {
			logger.Error("failed-evacuating", errors.New(response.Err))
			writeJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

func newPingHandler(logger lager.Logger, client driveradmin.DriverAdmin) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("handle-ping")
		logger.Info("start")
		defer logger.Info("end")

		env := driverhttp.EnvWithMonitor(logger, req.Context(), w)

		response := client.Ping(env)
		if response.Err != "" {
			logger.Error("failed-pinging", errors.New(response.Err))
			writeJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, jsonObj interface{}) {
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		panic("Unable to encode JSON: " + err.Error())
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(jsonBytes); err != nil {
		panic("Unable to write data: " + err.Error())
	}
}
