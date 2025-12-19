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
	"context"
	"errors"
	"fmt"
	"log/slog"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/abcxyz/pkg/logging"
)

// BigQuery provides a client and dataset identifiers.
type BigQuery struct {
	projectID string
	datasetID string
	client    *bigquery.Client
	logger    *slog.Logger
}

// CheckpointEntry is the shape of an entry to the checkpoint table.
type CheckpointEntry struct {
	deliveryID   string
	createdAt    string
	githubDomain string
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
		logger:    logging.FromContext(ctx),
	}, nil
}

// Close releases any resources held by the BigQuery client.
func (bq *BigQuery) Close() error {
	if err := bq.client.Close(); err != nil {
		return fmt.Errorf("failed to close BigQuery client: %w", err)
	}
	return nil
}

// Retrieve the latest checkpoint cursor value (deliveryID) in the checkpoint
// table. This is used by the retry service.
func (bq *BigQuery) RetrieveCheckpointID(ctx context.Context, checkpointTableID, githubDomain string) (string, error) {
	logger := logging.FromContext(ctx)
	var whereClause string
	var params []bigquery.QueryParameter

	if githubDomain == defaultGitHubDomain {
		whereClause = "WHERE t.github_instance_url IS NULL OR t.github_instance_url = ''"
	} else {
		whereClause = "WHERE t.github_instance_url = @githubDomain"
		params = []bigquery.QueryParameter{
			{
				Name:  "githubDomain",
				Value: githubDomain,
			},
		}
	}

	sqlQuery := fmt.Sprintf("SELECT t.delivery_id FROM `%s.%s.%s` AS t %s ORDER BY CAST(t.delivery_id AS INT64) DESC LIMIT 1",
		bq.projectID,
		bq.datasetID,
		checkpointTableID,
		whereClause,
	)

	logger.InfoContext(
		ctx,
		"reading checkpoing from BigQuery",
		"query",
		sqlQuery,
		"params",
		params,
	)
	q := bq.client.Query(sqlQuery)
	q.Parameters = params

	res, err := q.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to make read request to BigQuery: %w", err)
	}

	var rows []bigquery.Value
	nextErr := res.Next(&rows)

	if nextErr != nil {
		// If no checkpoint ID exists for this specific githubDomain, return an empty string.
		if errors.Is(nextErr, iterator.Done) {
			return "", nil
		}
		return "", fmt.Errorf("failed to iterate over query response: %w", nextErr)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("unexpected response from RetrieveCheckpointID : %s", rows)
	}

	checkpoint, ok := rows[0].(string)
	if !ok {
		return "", fmt.Errorf("failed to convert row value %v to string (got %T) for delivery_id", rows[0], rows[0])
	}

	return checkpoint, nil
}

// WriteCheckpointID writes the latest checkpoint that was successfully processed.
// This is used by the retry service to mark the last successfully processed event.
func (bq *BigQuery) WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt, githubDomain string) error {
	inserter := bq.client.Dataset(bq.datasetID).Table(checkpointTableID).Inserter()

	domainToInsert := githubDomain

	// If the domain is the default, insert an empty string ("") to handle records
	// from the default GitHub instance.
	if githubDomain == defaultGitHubDomain {
		domainToInsert = ""
	}

	// Create the CheckpointEntry, passing the raw string, relying on the BigQuery client
	// to convert the string to a TIMESTAMP during insertion.
	items := []*CheckpointEntry{
		{
			deliveryID:   deliveryID,
			createdAt:    createdAt,
			githubDomain: domainToInsert,
		},
	}

	if err := inserter.Put(ctx, items); err != nil {
		return fmt.Errorf("failed to execute WriteCheckpointID for deliveryID %s: %w", deliveryID, err)
	}

	return nil
}

// Check if an entry with a given delivery_id already exists in the events
// table, this attempts to prevent duplicate processing of events.
func (bq *BigQuery) DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error) {
	res, err := bq.makeCountQuery(ctx, eventsTableID, deliveryID)
	if err != nil {
		return false, fmt.Errorf("failed to execute DeliveryEventExists: %w", err)
	}

	return res > 0, nil
}

// Helper method to execute a count query for a given table by deliveryID and
// return the count.
func (bq *BigQuery) makeCountQuery(ctx context.Context, tableID, deliveryID string) (int64, error) {
	q := bq.client.Query(fmt.Sprintf("SELECT COUNT(1) FROM `%s.%s.%s` WHERE delivery_id = @deliveryID", bq.projectID, bq.datasetID, tableID))

	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "deliveryID",
			Value: deliveryID,
		},
	}

	// Execute the query.
	res, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to make read request to BigQuery: %w", err)
	}

	var rows []bigquery.Value
	if err := res.Next(&rows); err != nil {
		return 0, fmt.Errorf("failed to iterate over query response: %w", err)
	}

	if len(rows) == 0 {
		return 0, fmt.Errorf("unexpected response from querying %s: %s", tableID, rows)
	}

	count, ok := rows[0].(int64)
	if !ok {
		return 0, fmt.Errorf("failed to convert row value %v to int64 (got %T)", rows[0], rows[0])
	}

	return count, nil
}

// Save implements the ValueSaver interface for a CheckpointEntry. A random
// insertID is generated by the library to facilitate deduplication.
func (ce *CheckpointEntry) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"delivery_id":         ce.deliveryID,
		"created":             ce.createdAt,
		"github_instance_url": ce.githubDomain,
	}, "", nil
}
