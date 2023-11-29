// Copyright 2023 The Authors (see AUTHORS file)
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

package teeth

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-cmp/cmp"
)

const (
	testPullRequestEventsTable = "github_metrics_aggregator.pull_request_events"
	testEventsTable            = "github_metrics_aggregator.events"
	testLeechTable             = "github_metrics_aggregator.leech_status"
	testInvocationCommentTable = "github_metrics_aggregator.invocation_comment_status"
)

type fakeBigQueryClient struct {
	config *BQConfig
}

func DefaultFakeBigQueryClient() *fakeBigQueryClient {
	config := &BQConfig{
		PullRequestEventsTable:       testPullRequestEventsTable,
		EventsTable:                  testEventsTable,
		LeechStatusTable:             testLeechTable,
		InvocationCommentStatusTable: testInvocationCommentTable,
	}
	return &fakeBigQueryClient{
		config: config,
	}
}

func (f *fakeBigQueryClient) Config() *BQConfig {
	return f.config
}

func (f *fakeBigQueryClient) Query(q string) *bigquery.Query {
	return &bigquery.Query{
		QueryConfig: bigquery.QueryConfig{Q: q},
	}
}

func TestSetUpPublisherSourceQuery(t *testing.T) {
	t.Parallel()

	want := `-- Copyright 2023 The Authors (see AUTHORS file)
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
SELECT
  pull_request_events.delivery_id,
  delivery_events.pull_request_id,
  html_url AS pull_request_html_url,
  delivery_events.received,
  logs_uri,
  head_sha
FROM
  ` + "`" + testPullRequestEventsTable + "`" + ` AS pull_request_events
JOIN (
  SELECT
    delivery_id,
    received,
    logs_uri,
    SAFE.INT64(pull_request.id) AS pull_request_id,
    LAX_STRING(pull_request.url) AS pull_request_url,
    LAX_STRING(events.payload.workflow_run.head_sha) AS head_sha,
  FROM
    ` + "`" + testLeechTable + "`" + ` leech_status
  JOIN (
	  SELECT
      *
    FROM
      ` + "`" + testEventsTable + "`" + ` events,
      UNNEST(JSON_EXTRACT_ARRAY(events.payload.workflow_run.pull_requests)) AS pull_request
    WHERE
      received >= TIMESTAMP_ADD(CURRENT_TIMESTAMP(), INTERVAL -30 DAY)) AS events
  USING
    (delivery_id)) AS delivery_events
ON
  pull_request_events.id = delivery_events.pull_request_id
WHERE
  pull_request_events.id NOT IN (
  SELECT
    DISTINCT pull_request_id
  FROM
    ` + "`" + testInvocationCommentTable + "`" + ` invocation_comment_status)
  AND merged_at BETWEEN TIMESTAMP_ADD(CURRENT_TIMESTAMP(), INTERVAL -30 DAY)
  AND TIMESTAMP_ADD(CURRENT_TIMESTAMP(), INTERVAL -1 HOUR)
ORDER BY
  received, 
  pull_request_events.id ASC
`

	c := DefaultFakeBigQueryClient()
	q, err := SetUpPublisherSourceQuery(context.Background(), c)
	if err != nil {
		t.Errorf("SetUpPublisherSourceQuery returned unexpected error: %v", err)
	}
	if diff := cmp.Diff(want, q.QueryConfig.Q); diff != "" {
		t.Errorf("embedded source query mismatch  (-want +got):\n%s", diff)
	}
}
