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
	"testing"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/pkg/testutil"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	githubConfig := githubclient.Config{
		GitHubAppID:      "gh-app-123",
		GitHubPrivateKey: "private-key",
	}

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "missing_bucket_url",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				CheckpointTableID: "checkpoint-table-id",
				EventsTableID:     "events-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
			wantErr: `BUCKET_NAME is required`,
		},
		{
			name: "missing_checkpoint_table_id",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				BucketName:        "test-bucket-name",
				EventsTableID:     "events-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
			wantErr: `CHECKPOINT_TABLE_ID is required`,
		},
		{
			name: "missing_events_table_id",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				BucketName:        "test-bucket-name",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
			wantErr: `EVENTS_TABLE_ID is required`,
		},
		{
			name: "missing_dataset_id",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				BucketName:        "test-bucket-name",
				CheckpointTableID: "checkpoint-table-id",
				EventsTableID:     "events-table-id",
				ProjectID:         "test-project-id",
			},
			wantErr: `DATASET_ID is required`,
		},
		{
			name: "missing_project_id",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				BucketName:        "test-bucket-name",
				CheckpointTableID: "checkpoint-table-id",
				EventsTableID:     "events-table-id",
				DatasetID:         "test-dataset-id",
			},
			wantErr: `PROJECT_ID is required`,
		},
		{
			name: "success_fallback_bq_project_id",
			cfg: &Config{
				GitHub:            githubConfig,
				BucketName:        "test-bucket-name",
				CheckpointTableID: "checkpoint-table-id",
				EventsTableID:     "events-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
		},
		{
			name: "success",
			cfg: &Config{
				GitHub:            githubConfig,
				BigQueryProjectID: "test-bq-id",
				BucketName:        "test-bucket-name",
				CheckpointTableID: "checkpoint-table-id",
				EventsTableID:     "events-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			err := tc.cfg.Validate(ctx)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}
		})
	}
}
