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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	//nolint:gosec // This is a false positive for a variable name.
	serverWebhookSecret = "test-webhook-secret"
	serverProjectID     = "test-project-id"
	serverTopicID       = "test-topic-id"
)

func setupPubSubServer(ctx context.Context, t *testing.T, projectID, topicID string, opts ...pstest.ServerReactorOption) *grpc.ClientConn {
	t.Helper()

	// Create PubSub test server
	srv := pstest.NewServer(opts...)

	// Connect to the server without using TLS.
	conn, err := grpc.DialContext(ctx, srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("fail to connect to test pubsub server: %v", err)
	}

	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("fail to create test pubsub server client: %v", err)
	}

	// Create the test topic
	_, err = client.CreateTopic(ctx, topicID)
	if err != nil {
		t.Fatalf("failed to create test pubsub topic: %v", err)
	}

	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Fatalf("failed to cleanup test pubsub server: %v", err)
		}

		if err := conn.Close(); err != nil {
			t.Fatalf("failed to cleanup test pubsub client: %v", err)
		}
	})

	return conn
}

func TestHandleWebhook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testDataBasePath := path.Join("..", "..", "testdata")
	pubSubGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverTopicID)
	pubSubErrGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverTopicID, pstest.WithErrorInjection("Publish", codes.NotFound, "topic id not found"))

	cases := []struct {
		name                 string
		pubSubGRPCConn       *grpc.ClientConn
		payloadFile          string
		payloadType          string
		payloadWebhookSecret string
		messengerErr         string
		expStatusCode        int
		expRespBody          string
	}{
		{
			name:                 "success",
			pubSubGRPCConn:       pubSubGRPCConn,
			payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
			payloadType:          "pull_request",
			payloadWebhookSecret: serverWebhookSecret,
			expStatusCode:        http.StatusCreated,
			expRespBody:          successMessage,
		}, {
			name:                 "success_empty_payload",
			pubSubGRPCConn:       pubSubGRPCConn,
			payloadType:          "pull_request",
			payloadWebhookSecret: serverWebhookSecret,
			expStatusCode:        http.StatusBadRequest,
			expRespBody:          errNoPayload,
		}, {
			name:                 "invalid_signature",
			pubSubGRPCConn:       pubSubGRPCConn,
			payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
			payloadType:          "pull_request",
			payloadWebhookSecret: "not-valid",
			expStatusCode:        http.StatusUnauthorized,
			expRespBody:          errInvalidSignature,
		}, {
			name:                 "error_write_backend",
			pubSubGRPCConn:       pubSubErrGRPCConn,
			payloadFile:          path.Join(testDataBasePath, "pull_request.json"),
			payloadType:          "pull_request",
			payloadWebhookSecret: serverWebhookSecret,
			messengerErr:         "test backend error",
			expStatusCode:        http.StatusInternalServerError,
			expRespBody:          errWritingToBackend,
		},
	}

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

			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
			req.Header.Add(DeliveryIDHeader, "delivery-id")
			req.Header.Add(EventTypeHeader, tc.payloadType)
			req.Header.Add(SHA256SignatureHeader, fmt.Sprintf("sha256=%s", createSignature([]byte(tc.payloadWebhookSecret), payload)))

			resp := httptest.NewRecorder()

			cfg := &ServiceConfig{
				WebhookSecret: serverWebhookSecret,
				ProjectID:     serverProjectID,
				TopicID:       serverTopicID,
			}

			opts := []option.ClientOption{option.WithGRPCConn(tc.pubSubGRPCConn)}
			srv, err := NewServer(ctx, cfg, opts...)
			if err != nil {
				t.Fatalf("failed to create new server: %v", err)
			}

			srv.handleWebhook().ServeHTTP(resp, req)

			if resp.Code != tc.expStatusCode {
				t.Errorf("StatusCode got: %d want: %d", resp.Code, tc.expStatusCode)
			}

			if resp.Body.String() != tc.expRespBody {
				t.Errorf("ResponseBody got: %s want: %s", resp.Body.String(), tc.expRespBody)
			}
		})
	}
}

// createSignature creates a HMAC 256 signature for the test request payload.
func createSignature(key, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
