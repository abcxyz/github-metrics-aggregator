// Copyright 2025 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package relay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/pkg/logging"
)

type MockMessenger struct {
	CapturedMessage []byte
	CapturedAttrs   map[string]string
	SendErr         error
}

func (m *MockMessenger) Send(ctx context.Context, msg []byte, attrs map[string]string) error {
	m.CapturedMessage = msg
	m.CapturedAttrs = attrs
	return m.SendErr
}

func (m *MockMessenger) Close() error {
	return nil
}

func TestServer_handleRelay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		requestBody      string
		messengerErr     error
		wantStatusCode   int
		wantBodyContains string
		wantMessage      string // check captured message if empty skip check
	}{
		{
			name: "success",
			requestBody: func() string {
				payload := map[string]interface{}{
					"repository":   map[string]interface{}{"id": 12345, "full_name": "org/repo"},
					"organization": map[string]interface{}{"id": 67890, "login": "org"},
					"enterprise":   map[string]interface{}{"id": 11111, "name": "ent"},
				}
				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
				event := map[string]interface{}{
					"delivery_id": "delivery-1",
					"signature":   "sig-1",
					"received":    "2023-10-26T12:00:00Z",
					"event":       "push",
					"payload":     string(payloadBytes),
				}
				eventBytes, err := json.Marshal(event)
				if err != nil {
					t.Fatalf("failed to marshal event: %v", err)
				}
				data := make(map[string]interface{})
				data["data"] = eventBytes

				msg := map[string]interface{}{
					"message": map[string]interface{}{
						"data":       eventBytes, // This will be base64 encoded
						"attributes": map[string]string{},
						"messageId":  "msg-1",
					},
					"subscription": "sub-1",
				}
				b, err := json.Marshal(msg)
				if err != nil {
					t.Fatalf("failed to marshal message: %v", err)
				}
				return string(b)
			}(),
			wantStatusCode: http.StatusOK,
			wantMessage:    `{"delivery_id":"delivery-1","signature":"sig-1","received":"2023-10-26T12:00:00Z","event":"push","organization_id":"67890","organization_name":"org","repository_id":"12345","repository_name":"org/repo","enterprise_id":"11111","enterprise_name":"ent","payload":"{\"enterprise\":{\"id\":11111,\"name\":\"ent\"},\"organization\":{\"id\":67890,\"login\":\"org\"},\"repository\":{\"full_name\":\"org/repo\",\"id\":12345}}" }`,
		},
		{
			name: "malformed_request_body",
			requestBody: `
{
  "message": {
    "data": "invalid-base64",
`,
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "failed to enrich event",
		},
		{
			name: "malformed_event_json",
			requestBody: func() string {
				// Valid base64 but invalid JSON inside
				return `{"message":{"data":"e30=","attributes":{},"messageId":"msg-1"},"subscription":"sub-1"}`
			}(),
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "failed to enrich event",
		},
		{
			name: "messenger_error",
			requestBody: func() string {
				payload := map[string]interface{}{
					"repository":   map[string]interface{}{"id": 12345, "full_name": "org/repo"},
					"organization": map[string]interface{}{"id": 67890, "login": "org"},
					"enterprise":   map[string]interface{}{"id": 11111, "name": "ent"},
				}
				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
				event := map[string]interface{}{
					"delivery_id": "delivery-1",
					"signature":   "sig-1",
					"received":    "2023-10-26T12:00:00Z",
					"event":       "push",
					"payload":     string(payloadBytes),
				}
				eventBytes, err := json.Marshal(event)
				if err != nil {
					t.Fatalf("failed to marshal event: %v", err)
				}
				msg := map[string]interface{}{
					"message": map[string]interface{}{
						"data":       eventBytes,
						"attributes": map[string]string{},
						"messageId":  "msg-1",
					},
					"subscription": "sub-1",
				}
				b, err := json.Marshal(msg)
				if err != nil {
					t.Fatalf("failed to marshal message: %v", err)
				}
				return string(b)
			}(),
			messengerErr:   context.Canceled,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

			messenger := &MockMessenger{SendErr: tc.messengerErr}
			s := &Server{
				relayMessenger:  messenger,
				messageEnricher: NewDefaultMessageEnricher(),
			}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			s.handleRelay().ServeHTTP(rec, req)

			if got, want := rec.Code, tc.wantStatusCode; got != want {
				t.Errorf("StatusCode: got %d, want %d", got, want)
			}

			if tc.wantBodyContains != "" {
				if !strings.Contains(rec.Body.String(), tc.wantBodyContains) {
					t.Errorf("Body: got %q, want check to contain %q", rec.Body.String(), tc.wantBodyContains)
				}
			}

			if tc.wantMessage != "" {
				if messenger.CapturedMessage == nil {
					t.Error("CapturedMessage is nil, want message")
				} else {
					var got interface{}
					var want interface{}
					if err := json.Unmarshal(messenger.CapturedMessage, &got); err != nil {
						t.Fatalf("failed to unmarshal captured message: %v", err)
					}
					if err := json.Unmarshal([]byte(tc.wantMessage), &want); err != nil {
						t.Fatalf("failed to unmarshal want message: %v", err)
					}

					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("Message diff (-want, +got):\n%s", diff)
					}
				}
			}
		})
	}
}
