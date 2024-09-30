// Copyright 2024 The Authors (see AUTHORS file)
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

package artifact

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/abcxyz/github-metrics-aggregator/pkg/bq"
)

// sourceQuery is the driving BigQuery query that selects events
// that need to be processed.
const sourceQuery = `
SELECT
	delivery_id,
	JSON_VALUE(payload, "$.repository.full_name") repo_slug,
	JSON_VALUE(payload, "$.repository.name") repo_name,
	JSON_VALUE(payload, "$.repository.owner.login") org_name,
	JSON_VALUE(payload, "$.workflow_run.logs_url") logs_url,
	JSON_VALUE(payload, "$.workflow_run.actor.login") github_actor,
	JSON_VALUE(payload, "$.workflow_run.html_url") workflow_url,
	JSON_VALUE(payload, "$.workflow_run.id") workflow_run_id,
	JSON_VALUE(payload, "$.workflow_run.run_attempt") workflow_run_attempt,
	ARRAY(
		SELECT
			JSON_QUERY(pull_request, "$.number")
		FROM UNNEST(
			JSON_QUERY_ARRAY(payload, "$.workflow_run.pull_requests")
		) pull_request
	) pull_request_numbers
FROM {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.EventTableID}}{{.BT}}
WHERE
event = "workflow_run"
AND JSON_VALUE(payload, "$.workflow_run.status") = "completed"
AND delivery_id NOT IN (
SELECT
  delivery_id
FROM {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.ArtifactTableID}}{{.BT}}
)
LIMIT {{.BatchSize}}
`

type queryParameters struct {
	ProjectID       string
	DatasetID       string
	EventTableID    string
	ArtifactTableID string
	BatchSize       int
	BT              string
}

// makeQuery renders a string template representing the SQL query.
func makeQuery(client *bq.BigQuery, eventsTable, artifactTable string, batchSize int) (string, error) {
	tmpl, err := template.New("query").Parse(sourceQuery)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &queryParameters{
		ProjectID:       client.ProjectID,
		DatasetID:       client.DatasetID,
		EventTableID:    eventsTable,
		ArtifactTableID: artifactTable,
		BatchSize:       batchSize,
		BT:              "`",
	}); err != nil {
		return "", fmt.Errorf("failed to apply query template parameters: %w", err)
	}
	return sb.String(), nil
}
