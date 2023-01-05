// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/google/go-github/v48/github"
)

const serverWebhookSecret = "server-secret"

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
		return fmt.Errorf("failed to unmarshal TestMessager.Send event: %w", err)
	}
	t.event = event

	return nil
}

func TestHandleWebhook(t *testing.T) {
	t.Parallel()

	testDataBasePath := path.Join("..", "..", "integration", "data")

	cases := []struct {
		name                 string
		payloadFile          string
		payloadType          string
		payloadWebhookSecret string
		messagerErr          string
		expStatusCode        int
		expRespBody          string
	}{{
		name:                 "success",
		payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
		payloadType:          "pull_request",
		payloadWebhookSecret: serverWebhookSecret,
		expStatusCode:        http.StatusCreated,
		expRespBody:          "Ok",
	}, {
		name:                 "success_empty_payload",
		payloadType:          "pull_request",
		payloadWebhookSecret: serverWebhookSecret,
		expStatusCode:        http.StatusBadRequest,
		expRespBody:          "No payload received.",
	}, {
		name:                 "invalid_signature",
		payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
		payloadType:          "pull_request",
		payloadWebhookSecret: "not-valid",
		expStatusCode:        http.StatusBadRequest,
		expRespBody:          "Failed to validate webhook signature.",
	}, {
		name:                 "error_write_backend",
		payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
		payloadType:          "pull_request",
		payloadWebhookSecret: serverWebhookSecret,
		messagerErr:          "test backend error",
		expStatusCode:        http.StatusInternalServerError,
		expRespBody:          "Failed to write to backend.",
	}}

	for _, tc := range cases {
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

			signature := CreateSignature([]byte(tc.payloadWebhookSecret), payload)

			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
			req.Header.Add(github.DeliveryIDHeader, "delivery-id")
			req.Header.Add(github.EventTypeHeader, tc.payloadType)
			req.Header.Add(github.SHA256SignatureHeader, fmt.Sprintf("sha256=%s", signature))

			testMessager := &TestMessager{errMsg: tc.messagerErr}
			srv, err := NewRouter(context.Background(), serverWebhookSecret, testMessager)
			if err != nil {
				t.Fatalf("failed to create new server: %v", err)
			}

			respCode, respMsg, _ := srv.processWebhookRequest(req)

			if respCode != tc.expStatusCode {
				t.Errorf("StatusCode want: %d got: %d", tc.expStatusCode, respCode)
			}

			if respMsg != tc.expRespBody {
				t.Errorf("ResponseBody want: %s got: %s", tc.expRespBody, respMsg)
			}
		})
	}
}

// CreateSignature creates a HMAC 256 signature for the test request payload.
func CreateSignature(key, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
