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
			name: "missing_dataset_id",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
			wantErr: "DATASET_ID is required",
		},
		{
			name: "missing_events_table_id",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
			wantErr: "EVENTS_TABLE_ID is required",
		},
		{
			name: "missing_failure_events_table_id",
			cfg: &Config{
				BigQueryProjectID:   "test-big-query-project-id",
				DatasetID:           "test-dataset-id",
				EventsTableID:       "test-events-table-id",
				ProjectID:           "test-project-id",
				EventsTopicID:       "test-events-topic-id",
				DLQEventsTopicID:    "test-dlq-events-topic-id",
				GitHubWebhookSecret: "test-github-webhook-secret",
				RetryLimit:          1,
			},
			wantErr: "FAILURE_EVENTS_TABLE_ID is required",
		},
		{
			name: "missing_project_id",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
			wantErr: "PROJECT_ID is required",
		},
		{
			name: "missing_event_topic_id",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
			wantErr: "EVENTS_TOPIC_ID is required",
		},
		{
			name: "missing_dlq_event_topic_id",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
			wantErr: "DLQ_EVENTS_TOPIC_ID is required",
		},
		{
			name: "missing_webhook_secret",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				RetryLimit:           1,
			},
			wantErr: "GITHUB_WEBHOOK_SECRET is required",
		},
		{
			name: "missing_retry_limit",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
			},
			wantErr: "RETRY_LIMIT is required and must be greater than 0",
		},
		{
			name: "success",
			cfg: &Config{
				BigQueryProjectID:    "test-big-query-project-id",
				DatasetID:            "test-dataset-id",
				EventsTableID:        "test-events-table-id",
				FailureEventsTableID: "test-failure-events-table-id",
				ProjectID:            "test-project-id",
				EventsTopicID:        "test-events-topic-id",
				DLQEventsTopicID:     "test-dlq-events-topic-id",
				GitHubWebhookSecret:  "test-github-webhook-secret",
				RetryLimit:           1,
			},
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
