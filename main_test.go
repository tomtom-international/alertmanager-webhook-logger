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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/kami-zh/go-capturer"
	"github.com/prometheus/alertmanager/template"
)

func TestService(t *testing.T) {
	data := newAlerts()
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handler{
		Logger: log.NewNopLogger(),
	}
	http.Handle("/", &handler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestLogAlerts(t *testing.T) {
	alerts := newAlerts()
	out := capturer.CaptureStdout(func() {
		logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))

		err := logAlerts(alerts, logger)
		if err != nil {
			t.Errorf("error occurred during logging")
		}
	})

	var logMessage map[string]string

	decoder := json.NewDecoder(strings.NewReader(out))

	// message 1 parsed
	err := decoder.Decode(&logMessage)

	if err != nil {
		t.Errorf("invalid json receved for alert 1")
	}

	checkMap(t, logMessage, alerts.CommonAnnotations)
	checkMap(t, logMessage, alerts.CommonLabels)
	checkMap(t, logMessage, alerts.GroupLabels)
	checkString(t, logMessage, "receiver", alerts.Receiver)
	checkString(t, logMessage, "externalURL", alerts.ExternalURL)
	checkMap(t, logMessage, alerts.Alerts[0].Labels)
	checkMap(t, logMessage, alerts.Alerts[0].Annotations)

	checkString(t, logMessage, "status", alerts.Alerts[0].Status)
	checkString(t, logMessage, "startsAt", alerts.Alerts[0].StartsAt.Format(time.RFC3339))
	checkString(t, logMessage, "endsAt", alerts.Alerts[0].EndsAt.Format(time.RFC3339))
	checkString(t, logMessage, "generatorURL", alerts.Alerts[0].GeneratorURL)

	// message 2 parsed
	err = decoder.Decode(&logMessage)

	if err != nil {
		t.Errorf("invalid json receved for alert 2")
	}

	checkMap(t, logMessage, alerts.Alerts[1].Labels)
	checkMap(t, logMessage, alerts.Alerts[1].Annotations)
}

func checkMap(t *testing.T, logMessage map[string]string, dict map[string]string) {
	for k, v := range dict {
		checkString(t, logMessage, k, v)
	}
}

func checkString(t *testing.T, logMessage map[string]string, k string, v string) {
	if _, exists := logMessage[k]; !exists {
		t.Errorf("attribute %s:%s is not present", k, v)
	}

	if logMessage[k] != v {
		t.Errorf("attribute %s:%s has unexpcted value %s", k, v, logMessage[k])
	}
}

func newAlerts() template.Data {
	return template.Data{
		Alerts: template.Alerts{
			template.Alert{
				Status:       "critical",
				Annotations:  map[string]string{"a_key": "a_value"},
				Labels:       map[string]string{"l_key": "l_value"},
				StartsAt:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				EndsAt:       time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC),
				GeneratorURL: "file://generatorUrl",
			},
			template.Alert{
				Annotations: map[string]string{"a_key_warn": "a_value_warn"},
				Labels:      map[string]string{"l_key_warn": "l_value_warn"},
				Status:      "warning",
			},
		},
		CommonAnnotations: map[string]string{"ca_key": "ca_value"},
		CommonLabels:      map[string]string{"cl_key": "cl_value"},
		GroupLabels:       map[string]string{"gl_key": "gl_value"},
		ExternalURL:       "file://externalUrl",
		Receiver:          "test-receiver",
	}
}
