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
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// BigQuery provides a client and dataset identifiers.
type BigQuery struct {
	projectID string
	datasetID string
	client    *bigquery.Client
}

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
	JSON_VALUE(payload, "$.workflow_run.html_url") workflow_url
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

// NewBigQuery creates a new instance of a BigQuery client.
func NewBigQuery(ctx context.Context, projectID, datasetID string, opts ...option.ClientOption) (*BigQuery, error) {
	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new bigquery client: %w", err)
	}

	return &BigQuery{
		projectID: projectID,
		datasetID: datasetID,
		client:    client,
	}, nil
}

// Close releases any resources held by the BigQuery client.
func (bq *BigQuery) Close() error {
	if err := bq.client.Close(); err != nil {
		return fmt.Errorf("failed to close BigQuery client: %w", err)
	}
	return nil
}

// Query takes a queryString (assumed to be valid SQL) and executes it against
// BigQuery using the given client. The results are then mapped to a slice of T,
// where each row in the result is mapped to a struct of type T.
func Query[T any](ctx context.Context, bq *BigQuery, eventsTable, artifactTable string, batchSize int) ([]*T, error) {
	tmpl, err := template.New("query").Parse(sourceQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &queryParameters{
		ProjectID:       bq.projectID,
		DatasetID:       bq.datasetID,
		EventTableID:    eventsTable,
		ArtifactTableID: artifactTable,
		BatchSize:       batchSize,
		BT:              "`",
	}); err != nil {
		return nil, fmt.Errorf("failed to apply query template parameters: %w", err)
	}
	fmt.Println(sb.String())

	query := bq.client.Query(sb.String())
	job, err := query.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("query.Run failed: %w", err)
	}
	rows, err := job.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("job.Read failed: %w", err)
	}
	items := make([]*T, 0, rows.TotalRows)
	var t T
	for {
		err := rows.Next(&t)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("rows.Next failed: %w", err)
		}
		items = append(items, &t)
	}
	return items, nil
}

func Write(ctx context.Context, bq *BigQuery, tableID string, artifacts []*ArtifactRecord) error {
	if err := bq.client.Dataset(bq.datasetID).Table(tableID).Inserter().Put(ctx, artifacts); err != nil {
		return fmt.Errorf("failed to write to BigQuery: %w", err)
	}
	return nil
}
