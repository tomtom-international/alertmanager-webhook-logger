/*
 * Copyright (C) 2020 TomTom N.V. (www.tomtom.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	errorLog "log"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/alertmanager/template"
)

type handler struct {
	Logger log.Logger
}

func main() {
	address := flag.String("address", ":6725", "address and port of service")
	json := flag.Bool("json", true, "enable json logging")
	flag.Parse()

	lw := log.NewSyncWriter(os.Stdout)
	var logger log.Logger
	if *json {
		logger = log.NewJSONLogger(lw)
	} else {
		logger = log.NewLogfmtLogger(lw)
	}
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)

	http.Handle("/", &handler{
		Logger: logger,
	})
	if err := http.ListenAndServe(*address, nil); err != nil {
		errorLog.Fatalf("failed to start http server: %v", err)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var alerts template.Data
	err := json.NewDecoder(r.Body).Decode(&alerts)
	if err != nil {
		errorLog.Printf("cannot parse content because of %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = logAlerts(alerts, h.Logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func logAlerts(alerts template.Data, logger log.Logger) error {
	logger = logWith(alerts.CommonAnnotations, logger)
	logger = logWith(alerts.CommonLabels, logger)
	logger = logWith(alerts.GroupLabels, logger)
	for _, alert := range alerts.Alerts {
		alertLogger := logWith(alert.Labels, logger)
		alertLogger = logWith(alert.Annotations, alertLogger)

		err := alertLogger.Log("status", alert.Status, "startsAt", alert.StartsAt, "endsAt", alert.EndsAt, "generatorURL", alert.GeneratorURL, "externalURL", alerts.ExternalURL, "receiver", alerts.Receiver)
		if err != nil {
			return err
		}
	}

	return nil
}

func logWith(values map[string]string, logger log.Logger) log.Logger {
	for k, v := range values {
		logger = log.With(logger, k, v)
	}
	return logger
}
