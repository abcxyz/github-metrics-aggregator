// Copyright 2022 GitHub Metrics Aggregator authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/google/go-github/v48/github"
)

const (
	projectID    = "project"
	topicID      = "topic"
	serverSecret = "server-secret"
)

// TestMessager implements the Messager interface for testing.
type TestMessager struct {
	errMsg string
	event  Event
}

// Send is used for testing the HandleWebhook function.
func (t *TestMessager) Send(ctx context.Context, msg []byte) error {
	if len(t.errMsg) > 0 {
		return fmt.Errorf("TestMessager.Send: %v", t.errMsg)
	}

	var event Event
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}
	t.event = event

	return nil
}

func TestHandleWebhook(t *testing.T) {
	t.Parallel()

	testDataBasePath := path.Join("..", "..", "integration", "data")

	cases := []struct {
		name          string
		payloadFile   string
		payloadType   string
		payloadSecret string
		messagerErr   string
		expStatusCode int
		expRespBody   string
	}{{
		name:          "success",
		payloadFile:   path.Join(testDataBasePath, "pull_request.json"),
		payloadType:   "pull_request",
		payloadSecret: serverSecret,
		expStatusCode: http.StatusCreated,
	}, {
		name:          "success_empty_payload",
		payloadType:   "pull_request",
		payloadSecret: serverSecret,
		expStatusCode: http.StatusBadRequest,
		expRespBody:   "No payload received.",
	}, {
		name:          "invalid_signature",
		payloadFile:   path.Join(testDataBasePath, "pull_request.json"),
		payloadType:   "pull_request",
		payloadSecret: "not-valid",
		expStatusCode: http.StatusBadRequest,
		expRespBody:   "Failed to validate webhook signature.",
	}, {
		name:          "error_write_backend",
		payloadFile:   path.Join(testDataBasePath, "pull_request.json"),
		payloadType:   "pull_request",
		payloadSecret: serverSecret,
		messagerErr:   "test backend error",
		expStatusCode: http.StatusInternalServerError,
		expRespBody:   "Failed to write to backend.",
	}}

	for i, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var payload []byte
			var err error
			if len(tc.payloadFile) > 0 {
				payload, err = os.ReadFile(tc.payloadFile)
				if err != nil {
					t.Fatalf("failed to create payload from file: %v", err)
				}
			}

			signature := createSignature(tc.payloadSecret, payload)

			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
			req.Header.Add(github.DeliveryIDHeader, fmt.Sprintf("delivery-id-%d", i))
			req.Header.Add(github.EventTypeHeader, tc.payloadType)
			req.Header.Add(github.SHA256SignatureHeader, fmt.Sprintf("sha256=%s", signature))

			w := httptest.NewRecorder()

			os.Setenv("PROJECT_ID", projectID)
			os.Setenv("TOPIC_ID", topicID)
			os.Setenv("WEBHOOK_SECRET", serverSecret)

			config, err := NewConfig(context.Background())
			if err != nil {
				t.Fatalf("failed to create server config: %v", err)
			}

			tm := &TestMessager{errMsg: tc.messagerErr}
			srv, err := New(context.Background(), config, tm)
			if err != nil {
				t.Fatalf("failed to create new server: %v", err)
			}

			srv.HandleWebhook(w, req)

			if w.Result().StatusCode != tc.expStatusCode {
				t.Errorf("StatusCode want: %d got: %d", tc.expStatusCode, w.Result().StatusCode)
			}

			response, err := io.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if string(response) != tc.expRespBody {
				t.Errorf("ResponseBody want: %s got: %s", tc.expRespBody, string(response))
			}

		})
	}

}

func createSignature(key string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
