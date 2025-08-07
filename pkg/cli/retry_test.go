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

	"github.com/sethvargo/go-envconfig"
	"google.golang.org/api/option"

	"github.com/abcxyz/github-metrics-aggregator/pkg/retry"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
)

func TestRetryServerCommand(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

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
			name:   "invalid_config_github_app_id",
			env:    map[string]string{},
			expErr: `GITHUB_APP_ID is required`,
		},
		{
			name: "invalid_config_missing_github-private-key_and_kms_key_id",
			env: map[string]string{
				"GITHUB_APP_ID": "test-github-app-id",
			},
			expErr: `GITHUB_PRIVATE_KEY or GITHUB_PRIVATE_KEY_KMS_KEY_ID is required`,
		},
		{
			name: "invalid_config_bucket_url",
			env: map[string]string{
				"GITHUB_APP_ID":      "test-github-app-id",
				"GITHUB_PRIVATE_KEY": "test-github-private-key",
			},
			expErr: `BUCKET_NAME is required`,
		},
		{
			name: "invalid_config_checkpoint_table_id",
			env: map[string]string{
				"GITHUB_APP_ID":      "test-github-app-id",
				"GITHUB_PRIVATE_KEY": "test-github-private-key",
				"BUCKET_NAME":        "test-bucket-name",
			},
			expErr: `CHECKPOINT_TABLE_ID is required`,
		},
		{
			name: "invalid_config_events_tablet_id",
			env: map[string]string{
				"GITHUB_APP_ID":       "test-github-app-id",
				"GITHUB_PRIVATE_KEY":  "test-github-private-key",
				"BUCKET_NAME":         "test-bucket-name",
				"CHECKPOINT_TABLE_ID": "checkpoint-table-id",
			},
			expErr: `EVENTS_TABLE_ID is required`,
		},
		{
			name: "invalid_config_dataset_id",
			env: map[string]string{
				"GITHUB_APP_ID":       "test-github-app-id",
				"GITHUB_PRIVATE_KEY":  "test-github-private-key",
				"BUCKET_NAME":         "test-bucket-name",
				"CHECKPOINT_TABLE_ID": "checkpoint-table-id",
				"EVENTS_TABLE_ID":     "events-table-id",
			},
			expErr: `DATASET_ID is required`,
		},
		{
			name: "invalid_config_project_id",
			env: map[string]string{
				"GITHUB_APP_ID":       "test-github-app-id",
				"GITHUB_PRIVATE_KEY":  "test-github-private-key",
				"BUCKET_NAME":         "test-bucket-name",
				"CHECKPOINT_TABLE_ID": "checkpoint-table-id",
				"EVENTS_TABLE_ID":     "events-table-id",
				"DATASET_ID":          "dataset-id",
			},
			expErr: `PROJECT_ID is required`,
		},
		{
			name: "too_many_private_keys",
			env: map[string]string{
				"GITHUB_APP_ID":                 "test-github-app-id",
				"BIG_QUERY_PROJECT_ID":          "test-bq-id",
				"BUCKET_NAME":                   "test-bucket-name",
				"CHECKPOINT_TABLE_ID":           "checkpoint-table-id",
				"EVENTS_TABLE_ID":               "events-table-id",
				"DATASET_ID":                    "test-dataset-id",
				"PROJECT_ID":                    "test-project-id",
				"GITHUB_PRIVATE_KEY_KMS_KEY_ID": "test-kms-key-id",
				"GITHUB_PRIVATE_KEY":            "test-github-private-key",
			},
			expErr: `only one of GITHUB_PRIVATE_KEY, GITHUB_PRIVATE_KEY_KMS_KEY_ID is required`,
		},
		{
			name: "happy_path",
			env: map[string]string{
				"GITHUB_APP_ID":        "test-github-app-id",
				"GITHUB_PRIVATE_KEY":   "test-github-private-key",
				"BIG_QUERY_PROJECT_ID": "test-bq-id",
				"BUCKET_NAME":          "test-bucket-name",
				"CHECKPOINT_TABLE_ID":  "checkpoint-table-id",
				"EVENTS_TABLE_ID":      "events-table-id",
				"DATASET_ID":           "test-dataset-id",
				"PROJECT_ID":           "test-project-id",
			},
		},
		{
			name: "happy_path_with_kms-key-id",
			env: map[string]string{
				"GITHUB_APP_ID":                 "test-github-app-id",
				"BIG_QUERY_PROJECT_ID":          "test-bq-id",
				"BUCKET_NAME":                   "test-bucket-name",
				"CHECKPOINT_TABLE_ID":           "checkpoint-table-id",
				"EVENTS_TABLE_ID":               "events-table-id",
				"DATASET_ID":                    "test-dataset-id",
				"PROJECT_ID":                    "test-project-id",
				"GITHUB_PRIVATE_KEY_KMS_KEY_ID": "test-kms-key-id",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, done := context.WithCancel(ctx)
			defer done()

			var cmd RetryServerCommand
			cmd.testFlagSetOpts = []cli.Option{cli.WithLookupEnv(envconfig.MultiLookuper(
				envconfig.MapLookuper(tc.env),
				envconfig.MapLookuper(map[string]string{
					// Make the test choose a random port.
					"PORT": "0",
				}),
			).Lookup)}
			cmd.testGCSLockClientOptions = []option.ClientOption{
				// Disable auth lookup in these tests, since we don't actually call BQ.
				option.WithoutAuthentication(),
			}
			cmd.testDatastore = &retry.MockDatastore{}
			cmd.testGitHub = &retry.MockGitHub{}

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
