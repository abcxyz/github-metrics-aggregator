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

const (
	testAppID              = "test-app-id"
	testBigQueryID         = "test-bq-id"
	testBucketURL          = "test-bucket-url"
	testLockTTLClockSkewMS = 1000
	testLockTTLMinutes     = 1
	testProjectID          = "test-project-id"
	testWebhookID          = "test-webhook-id"
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
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				BucketURL:          testBucketURL,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
		},
		{
			name: "missing_app_id",
			cfg: &Config{
				BigQueryID:         testBigQueryID,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
			wantErr: `GITHUB_APP_ID is empty and requires a value`,
		},
		{
			name: "missing_bigquery_id",
			cfg: &Config{
				AppID:              testAppID,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
			wantErr: `BIG_QUERY_ID is empty and requires a value`,
		},
		{
			name: "missing_bucket_url",
			cfg: &Config{
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
			wantErr: `BUCKET_URL is empty and requires a value`,
		},
		{
			name: "invalid_lock_ttl_clock_skew_ms",
			cfg: &Config{
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				BucketURL:          testBucketURL,
				LockTTLClockSkewMS: -1,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
			wantErr: "LockTTLClockSkewMS must be a positive value, got: -1",
		},
		{
			name: "invalid_lock_ttl_clock_skew_ms",
			cfg: &Config{
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				BucketURL:          testBucketURL,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     -1,
				ProjectID:          testProjectID,
				WebhookID:          testWebhookID,
			},
			wantErr: "LockTTLMinutes must be a positive value, got: -1",
		},
		{
			name: "missing_project_id",
			cfg: &Config{
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				BucketURL:          testBucketURL,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				WebhookID:          testWebhookID,
			},
			wantErr: `PROJECT_ID is empty and requires a value`,
		},
		{
			name: "missing_webhook_id",
			cfg: &Config{
				AppID:              testAppID,
				BigQueryID:         testBigQueryID,
				BucketURL:          testBucketURL,
				LockTTLClockSkewMS: testLockTTLClockSkewMS,
				LockTTLMinutes:     testLockTTLMinutes,
				ProjectID:          testProjectID,
			},
			wantErr: `GITHUB_WEBHOOK_ID is empty and requires a value`,
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
