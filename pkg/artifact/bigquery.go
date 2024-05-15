package artifact

import (
	"context"
	"errors"
	"fmt"

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
const SourceQuery = `
SELECT
	delivery_id,
	JSON_VALUE(payload, "$.repository.full_name") repo_slug,
	JSON_VALUE(payload, "$.repository.name") repo_name,
	JSON_VALUE(payload, "$.repository.owner.login") org_name,
	JSON_VALUE(payload, "$.workflow_run.logs_url") logs_url,
	JSON_VALUE(payload, "$.workflow_run.actor.login") github_actor,
	JSON_VALUE(payload, "$.workflow_run.html_url") workflow_url
FROM ` + "`%s`" + `
WHERE
event = "workflow_run"
AND JSON_VALUE(payload, "$.workflow_run.status") = "completed"
AND delivery_id NOT IN (
SELECT
  delivery_id
FROM ` + "`%s`" + `)
LIMIT %d
`

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

// formatGoogleSQL formats the qualified table name in GoogleSQL syntax.
// i.e. "<project>.<dataset>.<table>".
func formatGoogleSQL(projectID, datasetID, tableID string) string {
	return fmt.Sprintf("%s.%s.%s", projectID, datasetID, tableID)
}

// Query takes a queryString (assumed to be valid SQL) and executes it against
// BigQuery using the given client. The results are then mapped to a slice of T,
// where each row in the result is mapped to a struct of type T.
func Query[T any](ctx context.Context, bq *BigQuery, queryString string) ([]*T, error) {
	query := bq.client.Query(queryString)
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

func Write(ctx context.Context, bq *BigQuery, tableID string, artifacts []ArtifactRecord) error {
	if err := bq.client.Dataset(bq.datasetID).Table(tableID).Inserter().Put(ctx, artifacts); err != nil {
		return fmt.Errorf("failed to write to BigQuery: %w", err)
	}
	return nil
}
