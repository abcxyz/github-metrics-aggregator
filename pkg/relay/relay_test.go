package relay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-cmp/cmp"
)

type MockMessenger struct {
	CapturedMessage []byte
	SendErr         error
}

func (m *MockMessenger) Send(ctx context.Context, msg []byte) error {
	m.CapturedMessage = msg
	return m.SendErr
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
				payloadBytes, _ := json.Marshal(payload)
				event := map[string]interface{}{
					"delivery_id": "delivery-1",
					"signature":   "sig-1",
					"received":    "2023-10-26T12:00:00Z",
					"event":       "push",
					"payload":     string(payloadBytes),
				}
				eventBytes, _ := json.Marshal(event)
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
				b, _ := json.Marshal(msg)
				return string(b)
			}(),
			wantStatusCode: http.StatusOK,
			wantMessage:    `{"delivery_id":"delivery-1","signature":"sig-1","received":"2023-10-26T12:00:00Z","event":"push","organization_id":"67890","organization_name":"org","repository_id":"12345","repository_name":"org/repo","enterprise_id":"11111","enterprise_name":"ent"}`,
		},
		{
			name: "malformed_request_body",
			requestBody: `
{
  "message": {
    "data": "invalid-base64",
`,
			wantStatusCode:   http.StatusBadRequest,
			wantBodyContains: "failed to decode request body",
		},
		{
			name: "malformed_event_json",
			requestBody: func() string {
				// Valid base64 but invalid JSON inside
				return `{"message":{"data":"e30=","attributes":{},"messageId":"msg-1"},"subscription":"sub-1"}`
			}(),
			wantStatusCode:   http.StatusBadRequest,
			wantBodyContains: "failed to decode event payload",
		},
		{
			name: "messenger_error",
			requestBody: func() string {
				payload := map[string]interface{}{
					"repository":   map[string]interface{}{"id": 12345, "full_name": "org/repo"},
					"organization": map[string]interface{}{"id": 67890, "login": "org"},
					"enterprise":   map[string]interface{}{"id": 11111, "name": "ent"},
				}
				payloadBytes, _ := json.Marshal(payload)
				event := map[string]interface{}{
					"delivery_id": "delivery-1",
					"signature":   "sig-1",
					"received":    "2023-10-26T12:00:00Z",
					"event":       "push",
					"payload":     string(payloadBytes),
				}
				eventBytes, _ := json.Marshal(event)
				msg := map[string]interface{}{
					"message": map[string]interface{}{
						"data":       eventBytes,
						"attributes": map[string]string{},
						"messageId":  "msg-1",
					},
					"subscription": "sub-1",
				}
				b, _ := json.Marshal(msg)
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
				relayMessenger: messenger,
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
