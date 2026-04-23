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

package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/abcxyz/pkg/renderer"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	//nolint:gosec // This is a false positive for a variable name.
	serverGitHubWebhookSecret  = "test-github-webhook-secret"
	serverProjectID            = "test-project-id"
	serverEventsTopicID        = "test-events-topic-id"
	serverDLQEventsTopicID     = "test-dlq-events-topic-id"
	serverDatasetID            = "test-dataset-id"
	serverFailureEventsTableID = "test-failure-events-table-id"
)

func setupPubSubServer(ctx context.Context, t *testing.T, projectID, topicID string, opts ...pstest.ServerReactorOption) *grpc.ClientConn {
	t.Helper()

	// Create PubSub test server
	srv := pstest.NewServer(opts...)

	// Connect to the server without using TLS.
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	ctx := t.Context()

	testDataBasePath := path.Join("..", "..", "testdata")
	pubSubGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverEventsTopicID)
	pubSubErrGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverEventsTopicID, pstest.WithErrorInjection("Publish", codes.NotFound, "topic id not found"))
	dlqEventsPubSubGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverDLQEventsTopicID)
	dlqEventPubSubErrGRPCConn := setupPubSubServer(ctx, t, serverProjectID, serverDLQEventsTopicID, pstest.WithErrorInjection("Publish", codes.NotFound, "topic id not found"))

	cases := []struct {
		name                    string
		pubSubGRPCConn          *grpc.ClientConn
		dlqEventsPubSubGRPCConn *grpc.ClientConn
		payloadFile             string
		payloadBytes            []byte
		payloadType             string
		payloadWebhookSecret    string
		expStatusCode           int
		expRespBody             string
		datastoreOverride       Datastore
		wantPubSubMessage       string
	}{
		{
			name:                    "success",
			pubSubGRPCConn:          pubSubGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusCreated,
			expRespBody:             `{"status":"ok"}`,
			datastoreOverride:       &MockDatastore{},
		},
		{
			name:                    "success_empty_payload",
			pubSubGRPCConn:          pubSubGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusBadRequest,
			expRespBody:             `{"errors":["no payload received"]}`,
			datastoreOverride:       &MockDatastore{},
		},
		{
			name:                    "invalid_signature",
			pubSubGRPCConn:          pubSubGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    "not-valid",
			expStatusCode:           http.StatusUnauthorized,
			expRespBody:             `{"errors":["failed to validate webhook signature"]}`,
			datastoreOverride:       &MockDatastore{},
		},
		{
			name:                    "error_write_backend_failed_marshal",
			pubSubGRPCConn:          pubSubErrGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventPubSubErrGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusInternalServerError,
			expRespBody:             `{"errors":["failed to write to backend"]}`,
			datastoreOverride:       &MockDatastore{},
		},
		{
			name:                    "error_write_backend_failed_bq_failure_events_exceeds_retry_limit",
			pubSubGRPCConn:          pubSubErrGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusInternalServerError,
			expRespBody:             `{"errors":["failed to write to backend"]}`,
			datastoreOverride:       &MockDatastore{failureEventsExceedsRetryLimit: &failureEventsExceedsRetryLimitRes{res: false, err: errors.New("error")}},
		},
		{
			name:                    "error_write_backend_failed_bq_failure_events_exceeds_retry_limit_and_dlq_failed",
			pubSubGRPCConn:          pubSubErrGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventPubSubErrGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusInternalServerError,
			expRespBody:             `{"errors":["failed to write to backend"]}`,
			datastoreOverride:       &MockDatastore{failureEventsExceedsRetryLimit: &failureEventsExceedsRetryLimitRes{res: true}},
		},
		{
			name:                    "error_write_backend_failed_bq_failure_events_exceeds_retry_limit_dlq_success",
			pubSubGRPCConn:          pubSubErrGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusCreated,
			expRespBody:             `{"status":"ok"}`,
			datastoreOverride:       &MockDatastore{failureEventsExceedsRetryLimit: &failureEventsExceedsRetryLimitRes{res: true}},
		},
		{
			name:                    "error_write_backend_failed_bq_write_failure_event",
			pubSubGRPCConn:          pubSubErrGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadFile:             path.Join(testDataBasePath, "pull_request.json"),
			payloadType:             "pull_request",
			payloadWebhookSecret:    serverGitHubWebhookSecret,
			expStatusCode:           http.StatusInternalServerError,
			expRespBody:             `{"errors":["failed to write to backend"]}`,
			datastoreOverride:       &MockDatastore{},
		},
		{
			name:                    "enrichment_success",
			pubSubGRPCConn:          pubSubGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadBytes: func() []byte {
				payload := map[string]any{
					"repository":   map[string]any{"id": 12345, "full_name": "org/repo"},
					"organization": map[string]any{"id": 67890, "login": "org"},
					"enterprise":   map[string]any{"id": 11111, "name": "ent"},
				}
				b, err := json.Marshal(payload)
				if err != nil {
					panic(err)
				}
				return b
			}(),
			payloadType:          "push",
			payloadWebhookSecret: serverGitHubWebhookSecret,
			expStatusCode:        http.StatusCreated,
			expRespBody:          `{"status":"ok"}`,
			datastoreOverride:    &MockDatastore{},
			wantPubSubMessage: func() string {
				payload := map[string]any{
					"repository":   map[string]any{"id": 12345, "full_name": "org/repo"},
					"organization": map[string]any{"id": 67890, "login": "org"},
					"enterprise":   map[string]any{"id": 11111, "name": "ent"},
				}
				b, err := json.Marshal(payload)
				if err != nil {
					panic(err)
				}
				sig := createSignature([]byte(serverGitHubWebhookSecret), b)

				msg := map[string]any{
					"delivery_id":       "delivery-id",
					"signature":         "sha256=" + sig,
					"event":             "push",
					"organization_id":   "67890",
					"organization_name": "org",
					"repository_id":     "12345",
					"repository_name":   "org/repo",
					"enterprise_id":     "11111",
					"enterprise_name":   "ent",
					"payload":           string(b),
				}
				mb, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}
				return string(mb)
			}(),
		},
		{
			name:                    "enrichment_recover_org_from_installation_account",
			pubSubGRPCConn:          pubSubGRPCConn,
			dlqEventsPubSubGRPCConn: dlqEventsPubSubGRPCConn,
			payloadBytes: func() []byte {
				payload := map[string]any{
					"installation": map[string]any{
						"account": map[string]any{"type": "Organization", "id": 67890, "login": "org"},
					},
					"repository": map[string]any{"id": 12345, "full_name": "org/repo"},
				}
				b, err := json.Marshal(payload)
				if err != nil {
					panic(err)
				}
				return b
			}(),
			payloadType:          "installation_repositories",
			payloadWebhookSecret: serverGitHubWebhookSecret,
			expStatusCode:        http.StatusCreated,
			expRespBody:          `{"status":"ok"}`,
			datastoreOverride:    &MockDatastore{},
			wantPubSubMessage: func() string {
				payload := map[string]any{
					"installation": map[string]any{
						"account": map[string]any{"type": "Organization", "id": 67890, "login": "org"},
					},
					"repository": map[string]any{"id": 12345, "full_name": "org/repo"},
				}
				b, err := json.Marshal(payload)
				if err != nil {
					panic(err)
				}
				sig := createSignature([]byte(serverGitHubWebhookSecret), b)

				msg := map[string]any{
					"delivery_id":       "delivery-id",
					"signature":         "sha256=" + sig,
					"event":             "installation_repositories",
					"organization_id":   "67890",
					"organization_name": "org",
					"repository_id":     "12345",
					"repository_name":   "org/repo",
					"payload":           string(b),
				}
				mb, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}
				return string(mb)
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			var payload []byte
			var err error
			if len(tc.payloadFile) > 0 {
				payload, err = os.ReadFile(tc.payloadFile)
				if err != nil {
					t.Fatalf("failed to create payload from file: %v", err)
				}
			} else if len(tc.payloadBytes) > 0 {
				payload = tc.payloadBytes
			}

			var sub *pubsub.Subscription
			var pubsubClient *pubsub.Client
			if tc.wantPubSubMessage != "" {
				pubsubClient, err = pubsub.NewClient(ctx, serverProjectID, option.WithGRPCConn(tc.pubSubGRPCConn), option.WithoutAuthentication())
				if err != nil {
					t.Fatalf("failed to create pubsub client: %v", err)
				}
				sub, err = pubsubClient.CreateSubscription(ctx, "test-sub-"+tc.name, pubsub.SubscriptionConfig{
					Topic: pubsubClient.Topic(serverEventsTopicID),
				})
				if err != nil {
					t.Fatalf("failed to create subscription: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
			req.Header.Add(DeliveryIDHeader, "delivery-id")
			req.Header.Add(EventTypeHeader, tc.payloadType)
			req.Header.Add(SHA256SignatureHeader, fmt.Sprintf("sha256=%s", createSignature([]byte(tc.payloadWebhookSecret), payload)))

			resp := httptest.NewRecorder()

			cfg := &Config{
				DatasetID:            serverDatasetID,
				EventsTopicID:        serverEventsTopicID,
				DLQEventsTopicID:     serverDLQEventsTopicID,
				FailureEventsTableID: serverFailureEventsTableID,
				ProjectID:            serverProjectID,
				RetryLimit:           1,
				GitHubWebhookSecret:  serverGitHubWebhookSecret,
			}

			wco := &WebhookClientOptions{
				EventPubsubClientOpts:    []option.ClientOption{option.WithGRPCConn(tc.pubSubGRPCConn), option.WithoutAuthentication()},
				DLQEventPubsubClientOpts: []option.ClientOption{option.WithGRPCConn(tc.dlqEventsPubSubGRPCConn), option.WithoutAuthentication()},
				DatastoreClientOverride:  tc.datastoreOverride,
			}

			h, err := renderer.New(ctx, nil,
				renderer.WithDebug(true),
				renderer.WithOnError(func(err error) {
					t.Error(err)
				}))
			if err != nil {
				t.Fatal(err)
			}

			srv, err := NewServer(ctx, h, cfg, wco)
			if err != nil {
				t.Fatalf("failed to create new server: %v", err)
			}

			srv.handleWebhook().ServeHTTP(resp, req)

			if got, want := resp.Code, tc.expStatusCode; got != want {
				t.Errorf("expected %d to be %d", got, want)
			}

			if got, want := strings.TrimSpace(resp.Body.String()), tc.expRespBody; got != want {
				t.Errorf("expected %q to be %q", got, want)
			}

			if tc.wantPubSubMessage != "" {
				var gotMsg []byte
				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()
				err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
					gotMsg = m.Data
					m.Ack()
					cancel()
				})
				if err != nil && !errors.Is(err, context.Canceled) {
					t.Fatalf("failed to receive message: %v", err)
				}

				var got map[string]any
				var want map[string]any
				if err := json.Unmarshal(gotMsg, &got); err != nil {
					t.Fatalf("failed to unmarshal got message: %v", err)
				}
				if err := json.Unmarshal([]byte(tc.wantPubSubMessage), &want); err != nil {
					t.Fatalf("failed to unmarshal want message: %v", err)
				}

				delete(got, "received")
				delete(want, "received")

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("Message diff (-want, +got):\n%s", diff)
				}
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
