// Copyright 2023 Google LLC
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

package cli

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
	"github.com/sethvargo/go-envconfig"
)

func TestWebhookServerCommand(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))

	cases := []struct {
		name   string
		args   []string
		env    map[string]string
		expErr string
	}{
		{
			name:   "too_many_args",
			args:   []string{"foo"},
			expErr: `unexpected arguments: ["foo"]`,
		},
		{
			name:   "invalid_config_dataset_id",
			env:    map[string]string{},
			expErr: `DATASET_ID is required`,
		},
		{
			name: "invalid_config_events_table_id",
			env: map[string]string{
				"GITHUB_APP_ID": "test-app-id",
				"DATASET_ID":    "dataset-id",
			},
			expErr: `EVENTS_TABLE_ID is required`,
		},
		{
			name: "invalid_config_failure_events_table_id",
			env: map[string]string{
				"GITHUB_APP_ID":   "test-app-id",
				"DATASET_ID":      "dataset-id",
				"EVENTS_TABLE_ID": "events-table-id",
			},
			expErr: `FAILURE_EVENTS_TABLE_ID is required`,
		},
		{
			name: "invalid_config_project_id",
			env: map[string]string{
				"GITHUB_APP_ID":           "test-app-id",
				"DATASET_ID":              "dataset-id",
				"EVENTS_TABLE_ID":         "events-table-id",
				"FAILURE_EVENTS_TABLE_ID": "failure-events-table-id",
			},
			expErr: `PROJECT_ID is required`,
		},
		{
			name: "invalid_config_retry_limit",
			env: map[string]string{
				"GITHUB_APP_ID":           "test-app-id",
				"DATASET_ID":              "dataset-id",
				"EVENTS_TABLE_ID":         "events-table-id",
				"FAILURE_EVENTS_TABLE_ID": "failure-events-table-id",
				"PROJECT_ID":              "project-id",
			},
			expErr: `RETRY_LIMIT is required`,
		},
		{
			name: "invalid_config_topic_id",
			env: map[string]string{
				"GITHUB_APP_ID":           "test-app-id",
				"DATASET_ID":              "dataset-id",
				"EVENTS_TABLE_ID":         "events-table-id",
				"FAILURE_EVENTS_TABLE_ID": "failure-events-table-id",
				"PROJECT_ID":              "project-id",
				"RETRY_LIMIT":             "1",
			},
			expErr: `TOPIC_ID is required`,
		},
		{
			name: "invalid_config_webhook_secret",
			env: map[string]string{
				"GITHUB_APP_ID":           "test-app-id",
				"DATASET_ID":              "dataset-id",
				"EVENTS_TABLE_ID":         "events-table-id",
				"FAILURE_EVENTS_TABLE_ID": "failure-events-table-id",
				"PROJECT_ID":              "project-id",
				"RETRY_LIMIT":             "1",
				"TOPIC_ID":                "topic-id",
			},
			expErr: `WEBHOOK_SECRET is required`,
		},
		{
			name: "happy_path",
			env: map[string]string{
				"GITHUB_APP_ID":           "test-app-id",
				"DATASET_ID":              "dataset-id",
				"EVENTS_TABLE_ID":         "events-table-id",
				"FAILURE_EVENTS_TABLE_ID": "failure-events-table-id",
				"PROJECT_ID":              "project-id",
				"RETRY_LIMIT":             "1",
				"TOPIC_ID":                "topic-id",
				"WEBHOOK_SECRET":          "webhook-secret",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, done := context.WithCancel(ctx)
			defer done()

			var cmd WebhookServerCommand
			cmd.testFlagSetOpts = []cli.Option{cli.WithLookupEnv(envconfig.MultiLookuper(
				envconfig.MapLookuper(tc.env),
				envconfig.MapLookuper(map[string]string{
					// Make the test choose a random port.
					"PORT": "0",
				}),
			).Lookup)}

			_, _, _ = cmd.Pipe()

			srv, mux, err := cmd.RunUnstarted(ctx, tc.args)
			if diff := testutil.DiffErrString(err, tc.expErr); diff != "" {
				t.Fatal(diff)
			}
			if err != nil {
				return
			}

			serverCtx, serverDone := context.WithCancel(ctx)
			defer serverDone()
			go func() {
				if err := srv.StartHTTPHandler(serverCtx, mux); err != nil {
					t.Error(err)
				}
			}()

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			uri := "http://" + srv.Addr() + "/healthz"
			req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if got, want := resp.StatusCode, http.StatusOK; got != want {
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				t.Errorf("expected status code %d to be %d: %s", got, want, string(b))
			}
		})
	}
}
