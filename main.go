package main

import (
	"encoding/json"
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
	lw := log.NewSyncWriter(os.Stdout)
	logger := log.With(log.NewJSONLogger(lw), "timestamp", log.DefaultTimestampUTC)

	http.Handle("/", &handler{
		Logger: logger,
	})
	http.ListenAndServe(":6725", nil)
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
		logger = logWith(alert.Labels, logger)
		logger = logWith(alert.Annotations, logger)

		err := logger.Log("status", alert.Status, "startsAt", alert.StartsAt, "endsAt", alert.EndsAt, "generatorURL", alert.GeneratorURL, "externalURL", alerts.ExternalURL, "receiver", alerts.Receiver)
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
