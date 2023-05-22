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

package leech

import (
	"testing"

	"github.com/abcxyz/pkg/testutil"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "success",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
		},
		{
			name: "success_with_secrets",
			cfg: &Config{
				BatchSize:              10,
				EventsProjectID:        "test-events-project-id",
				EventsTable:            "test-events-table",
				GitHubAppIDSecret:      "test-app-id-secret",
				GitHubInstallIDSecret:  "test-install-id-secret",
				GitHubPrivateKeySecret: "test-private-key-secret",
				JobName:                "test-job-name",
				LeechProjectID:         "test-leech-project-id",
				LeechTable:             "test-leech-table",
				LogsBucketName:         "test-logs-bucket",
			},
		},
		{
			name: "missing_batch_size",
			cfg: &Config{
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "BATCH_SIZE must be a positive integer",
		},
		{
			name: "missing_app_id",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "GITHUB_APP_ID or GITHUB_APP_ID_SECRET is required",
		},
		{
			name: "missing_install_id",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "GITHUB_INSTALL_ID or GITHUB_INSTALL_ID_SECRET is required",
		},
		{
			name: "missing_private_key",
			cfg: &Config{
				BatchSize:       10,
				EventsProjectID: "test-events-project-id",
				EventsTable:     "test-events-table",
				GitHubAppID:     "test-app-id",
				GitHubInstallID: "test-install-id",
				JobName:         "test-job-name",
				LeechProjectID:  "test-leech-project-id",
				LeechTable:      "test-leech-table",
				LogsBucketName:  "test-logs-bucket",
			},
			wantErr: "GITHUB_PRIVATE_KEY or GITHUB_PRIVATE_KEY_SECRET is required",
		},
		{
			name: "missing_logs_bucket",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
			},
			wantErr: "LOGS_BUCKET_NAME is required",
		},
		{
			name: "missing_events_project_id",
			cfg: &Config{
				BatchSize:        10,
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "EVENTS_PROJECT_ID is required",
		},
		{
			name: "missing_events_table",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "EVENTS_TABLE is required",
		},
		{
			name: "missing_job_name",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				LeechProjectID:   "test-leech-project-id",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "JOB_NAME is required",
		},
		{
			name: "missing_leech_project_id",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechTable:       "test-leech-table",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "LEECH_PROJECT_ID is required",
		},
		{
			name: "missing_leech_table",
			cfg: &Config{
				BatchSize:        10,
				EventsProjectID:  "test-events-project-id",
				EventsTable:      "test-events-table",
				GitHubAppID:      "test-app-id",
				GitHubInstallID:  "test-install-id",
				GitHubPrivateKey: "test-private-key",
				JobName:          "test-job-name",
				LeechProjectID:   "test-leech-project-id",
				LogsBucketName:   "test-logs-bucket",
			},
			wantErr: "LEECH_TABLE is required",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}
		})
	}
}
