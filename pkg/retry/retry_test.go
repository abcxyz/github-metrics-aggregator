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

package retry

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v61/github"
	"github.com/sethvargo/go-gcslock"

	"github.com/abcxyz/pkg/renderer"
)

func TestHandleRetry(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	cases := []struct {
		name                    string
		expStatusCode           int
		expRespBody             string
		datastoreClientOverride Datastore
		gcsLockClientOverride   gcslock.Lockable
		githubOverride          GitHubSource
	}{
		{
			name:          "held_lock",
			expStatusCode: http.StatusOK,
			expRespBody:   `{"status":"ok"}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{
					err: gcslock.NewLockHeldError(1),
				},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](101),
							StatusCode: toPtr(http.StatusOK),
						},
					},
					res: &github.Response{},
				},
			},
		},
		{
			name:          "error_lock",
			expStatusCode: http.StatusInternalServerError,
			expRespBody:   `{"errors":["failed to acquire google cloud storage lock"]}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{
					err: errors.New("error"),
				},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](101),
							StatusCode: toPtr(http.StatusOK),
						},
					},
					res: &github.Response{},
				},
			},
		},
		{
			name:          "retrieve_checkpoint_failure",
			expStatusCode: http.StatusInternalServerError,
			expRespBody:   `{"errors":["failed to retrieve checkpoint"]}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{err: errors.New("error")},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](101),
							StatusCode: toPtr(http.StatusOK),
						},
					},
					res: &github.Response{},
				},
			},
		},
		{
			name:          "github_list_deliveries_failure",
			expStatusCode: http.StatusInternalServerError,
			expRespBody:   http.StatusText(http.StatusInternalServerError),
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{err: errors.New("error")},
			},
		},
		{
			name:          "github_list_deliveries_empty",
			expStatusCode: http.StatusAccepted,
			expRespBody:   `{"status":"accepted"}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{},
					res:        &github.Response{},
				},
			},
		},
		{
			name:          "github_redeliver_event_failure_big_query_entry_not_exists",
			expStatusCode: http.StatusInternalServerError,
			expRespBody:   http.StatusText(http.StatusInternalServerError),
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{err: errors.New("error")},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](1),
							StatusCode: toPtr(http.StatusInternalServerError),
							GUID:       toPtr("guid"),
							Event:      toPtr("event"),
						},
					},
					res: &github.Response{},
				},
				redeliverEvent: &redeliverEventRes{err: errors.New("error")},
			},
		},
		{
			name:          "github_redeliver_event_failure_big_query_entry_exists",
			expStatusCode: http.StatusAccepted,
			expRespBody:   `{"status":"accepted"}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{res: true},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](1),
							StatusCode: toPtr(http.StatusInternalServerError),
							GUID:       toPtr("guid"),
							Event:      toPtr("event"),
						},
					},
					res: &github.Response{},
				},
				redeliverEvent: &redeliverEventRes{err: errors.New("error")},
			},
		},
		{
			name:          "github_redeliver_event_failure",
			expStatusCode: http.StatusInternalServerError,
			expRespBody:   http.StatusText(http.StatusInternalServerError),
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{res: false},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](1),
							StatusCode: toPtr(http.StatusInternalServerError),
							GUID:       toPtr("guid"),
							Event:      toPtr("event"),
						},
					},
					res: &github.Response{},
				},
				redeliverEvent: &redeliverEventRes{err: errors.New("error")},
			},
		},
		{
			name:          "success",
			expStatusCode: http.StatusAccepted,
			expRespBody:   `{"status":"accepted"}`,
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				acquire: &acquireRes{},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{
						{
							ID:         toPtr[int64](101),
							StatusCode: toPtr(http.StatusOK),
						},
					},
					res: &github.Response{},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h, err := renderer.New(ctx, nil,
				renderer.WithDebug(true),
				renderer.WithOnError(func(err error) {
					t.Error(err)
				}))
			if err != nil {
				t.Fatal(err)
			}

			srv, err := NewServer(ctx, h, &Config{}, &RetryClientOptions{
				DatastoreClientOverride: tc.datastoreClientOverride,
				GCSLockClientOverride:   tc.gcsLockClientOverride,
				GitHubOverride:          tc.githubOverride,
			})
			if err != nil {
				t.Fatalf("failed to create new server: %v", err)
			}

			var payload []byte
			req := httptest.NewRequest(http.MethodPost, "/retry", bytes.NewReader(payload))

			resp := httptest.NewRecorder()

			srv.handleRetry().ServeHTTP(resp, req)

			if resp.Code != tc.expStatusCode {
				t.Errorf("StatusCode got: %d want: %d", resp.Code, tc.expStatusCode)
			}

			if strings.TrimSpace(resp.Body.String()) != tc.expRespBody {
				t.Errorf("ResponseBody got: %s want: %s", resp.Body.String(), tc.expRespBody)
			}
		})
	}
}

// toPtr is a helper function to convert a type to a pointer of that same type.
func toPtr[T any](i T) *T {
	return &i
}
