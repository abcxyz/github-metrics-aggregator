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

package retry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v61/github"
	"github.com/sethvargo/go-gcslock"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
)

func TestExecuteJob(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	cases := []struct {
		name                    string
		expErr                  string
		datastoreClientOverride Datastore
		gcsLockClientOverride   gcslock.Lockable
		githubOverride          GitHubSource
	}{
		{
			name:   "held_lock",
			expErr: "",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return gcslock.NewLockHeldError(1)
				},
				CloseFn: func(context.Context) error {
					return nil
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
			name:   "error_lock",
			expErr: "failed to acquire gcs lock: error",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return errors.New("error")
				},
				CloseFn: func(context.Context) error {
					return nil
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
			name:   "retrieve_checkpoint_failure",
			expErr: "failed to retrieve checkpoint: error",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{err: errors.New("error")},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
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
			name:   "github_list_deliveries_failure",
			expErr: "failed to list deliveries: error",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
				},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{err: errors.New("error")},
			},
		},
		{
			name:   "github_list_deliveries_empty",
			expErr: "",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
				},
			},
			githubOverride: &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					deliveries: []*github.HookDelivery{},
					res:        &github.Response{},
				},
			},
		},
		{
			name:   "github_redeliver_event_failure_big_query_entry_not_exists",
			expErr: "failed to check if delivery event exists: error",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{err: errors.New("error")},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
				},
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
			name:   "github_redeliver_event_failure_big_query_entry_exists",
			expErr: "",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{res: true},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
				},
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
			name:   "github_redeliver_event_failure",
			expErr: "failed to redeliver event: error",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
				deliveryEventExists:  &deliveryEventExistsRes{res: false},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
				},
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
			name:   "success",
			expErr: "",
			datastoreClientOverride: &MockDatastore{
				retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
			},
			gcsLockClientOverride: &MockLock{
				AcquireFn: func(context.Context, time.Duration) error {
					return nil
				},
				CloseFn: func(context.Context) error {
					return nil
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ExecuteJob(ctx, &Config{}, &RetryClientOptions{
				DatastoreClientOverride: tc.datastoreClientOverride,
				GCSLockClientOverride:   tc.gcsLockClientOverride,
				GitHubOverride:          tc.githubOverride,
			})

			if tc.expErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tc.expErr)
				} else if diff := cmp.Diff(err.Error(), tc.expErr); diff != "" {
					t.Errorf("error diff (-want, +got):\n%s", diff)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %q", err)
			}
		})
	}
}

func TestExecuteJob_TokenRefresh(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	var clientCreateCount int
	clientCreator := func(ctx context.Context, cfg *githubclient.Config) (GitHubSource, error) {
		clientCreateCount++
		if clientCreateCount == 1 {
			return &MockGitHub{
				listDeliveries: &listDeliveriesRes{
					err: &github.ErrorResponse{
						Response: &http.Response{StatusCode: 401},
						Message:  "Bad credentials",
					},
				},
			}, nil
		}
		return &MockGitHub{
			listDeliveries: &listDeliveriesRes{
				deliveries: []*github.HookDelivery{
					{ID: toPtr[int64](102), StatusCode: toPtr(http.StatusOK)},
				},
				res: &github.Response{Cursor: ""},
			},
		}, nil
	}

	err := ExecuteJob(ctx, &Config{}, &RetryClientOptions{
		DatastoreClientOverride: &MockDatastore{
			retrieveCheckpointID: &retrieveCheckpointIDRes{res: "checkpoint-id"},
		},
		GCSLockClientOverride: &MockLock{
			AcquireFn: func(context.Context, time.Duration) error { return nil },
			CloseFn:   func(context.Context) error { return nil },
		},
		GitHubClientCreator: clientCreator,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientCreateCount != 2 {
		t.Errorf("expected 2 client creations (initial + refresh), got %d", clientCreateCount)
	}
}

// toPtr is a helper function to convert a type to a pointer of that same type.
func toPtr[T any](i T) *T {
	return &i
}
