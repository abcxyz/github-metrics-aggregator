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
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
				WebhookID:         "test-webhook-id",
			},
		},
		{
			name: "missing_app_id",
			cfg: &Config{
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
				WebhookID:         "test-webhook-id",
			},
			wantErr: `GITHUB_APP_ID is required`,
		},
		{
			name: "missing_bucket_url",
			cfg: &Config{
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
				WebhookID:         "test-webhook-id",
			},
			wantErr: `BUCKET_URL is required`,
		},
		{
			name: "missing_checkpoint_table_id",
			cfg: &Config{
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
				WebhookID:         "test-webhook-id",
			},
			wantErr: `CHECKPOINT_TABLE_ID is required`,
		},
		{
			name: "missing_dataset_id",
			cfg: &Config{
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				CheckpointTableID: "checkpoint-table-id",
				ProjectID:         "test-project-id",
				WebhookID:         "test-webhook-id",
			},
			wantErr: `DATASET_ID is required`,
		},
		{
			name: "missing_project_id",
			cfg: &Config{
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				WebhookID:         "test-webhook-id",
			},
			wantErr: `PROJECT_ID is required`,
		},
		{
			name: "missing_webhook_id",
			cfg: &Config{
				AppID:             "test-app-id",
				BigQueryProjectID: "test-bq-id",
				BucketURL:         "test-bucket-url",
				CheckpointTableID: "checkpoint-table-id",
				DatasetID:         "test-dataset-id",
				ProjectID:         "test-project-id",
			},
			wantErr: `GITHUB_WEBHOOK_ID is required`,
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
