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

	checkMap(t, out, alerts.CommonAnnotations)
	checkMap(t, out, alerts.CommonLabels)
	checkMap(t, out, alerts.GroupLabels)
	checkString(t, out, "receiver", alerts.Receiver)
	checkString(t, out, "externalURL", alerts.ExternalURL)
	checkMap(t, out, alerts.Alerts[0].Labels)
	checkMap(t, out, alerts.Alerts[0].Annotations)

	checkString(t, out, "status", alerts.Alerts[0].Status)
	checkString(t, out, "startsAt", alerts.Alerts[0].StartsAt.Format(time.RFC3339))
	checkString(t, out, "endsAt", alerts.Alerts[0].EndsAt.Format(time.RFC3339))
	checkString(t, out, "generatorURL", alerts.Alerts[0].GeneratorURL)
}

func checkMap(t *testing.T, out string, dict map[string]string) {
	for k, v := range dict {
		checkString(t, out, k, v)
	}
}

func checkString(t *testing.T, out string, k string, v string) {
	if !strings.Contains(out, "\""+k+"\":\""+v+"\"") {
		t.Errorf("attribute %s:%s is missing in %s", k, v, out)
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
				Status: "warning",
			},
		},
		CommonAnnotations: map[string]string{"ca_key": "ca_value"},
		CommonLabels:      map[string]string{"cl_key": "cl_value"},
		GroupLabels:       map[string]string{"gl_key": "gl_value"},
		ExternalURL:       "file://externalUrl",
		Receiver:          "test-receiver",
	}
}
