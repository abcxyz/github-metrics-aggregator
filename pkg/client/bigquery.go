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

// Package client encapsulates all clients and associated helper methods to interact with other services.
package client

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

// CheckpointEntry is the shape of an entry to the checkpoint table.
type CheckpointEntry struct {
	deliveryID string
	createdAt  string
}

// FailureEventEntry is the shape of an entry to the failure_events table.
type FailureEventEntry struct {
	deliveryID string
	createdAt  string
}

// NewBigQueryClient creates a new instance of a BigQuery client.
func NewBigQueryClient(ctx context.Context, projectID, datasetID string, opts ...option.ClientOption) (*BigQuery, error) {
	client, err := bigquery.NewClient(ctx, projectID)
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
func (bq *BigQuery) Close() {
	bq.client.Close()
}

// Check if an entry with a given delivery_id already exists in the events table, this attempts to prevent duplicate processing of events.
// This is used by the webhook service.
func (bq *BigQuery) DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error) {
	res, err := bq.makeCountQuery(ctx, fmt.Sprintf("SELECT COUNT(1) FROM `%s.%s.%s` WHERE delivery_id = '%q'",
		bq.projectID,
		bq.datasetID,
		eventsTableID,
		deliveryID,
	))
	if err != nil {
		return false, fmt.Errorf("failed to execute DeliveryEventExists: %w", err)
	}

	return res > 0, nil
}

// Check if the number of entries with a given delivery_id in the failure-events table exceeds the retry limit.
// This is used by the webhook service.
func (bq *BigQuery) FailureEventsExceedsRetryLimit(ctx context.Context, failureEventTableID, deliveryID string, maxRetry int) (bool, error) {
	res, err := bq.makeCountQuery(ctx, fmt.Sprintf("SELECT COUNT(1) FROM `%s.%s.%s` WHERE delivery_id = '%q'",
		bq.projectID,
		bq.datasetID,
		failureEventTableID,
		deliveryID,
	))
	if err != nil {
		return false, fmt.Errorf("failed to execute FailureEventsExceedsRetryLimit: %w", err)
	}

	return res >= maxRetry, nil
}

// Write a failure event entry if there is a failure in processing the event.
// This is used by the webhook service.
func (bq *BigQuery) WriteFailureEvent(ctx context.Context, failureEventTableID, deliveryID, createdAt string) error {
	inserter := bq.client.Dataset(bq.datasetID).Table(failureEventTableID).Inserter()
	items := []*FailureEventEntry{
		// FailureEventEntry implements the ValueSaver interface.
		{deliveryID: deliveryID, createdAt: createdAt},
	}
	if err := inserter.Put(ctx, items); err != nil {
		return fmt.Errorf("failed to execute WriteFailureEvent: %w", err)
	}

	return nil
}

// Retrieve the latest checkpoint cursor value (deliveryID) in the checkpoint table.
// This is used by the retry service.
func (bq *BigQuery) RetrieveCheckpointID(ctx context.Context, checkpointTableID string) (string, error) {
	// Construct a query.
	q := bq.client.Query(fmt.Sprintf("SELECT delivery_id FROM `%s.%s.%s` ORDER BY created DESC LIMIT 1",
		bq.projectID,
		bq.datasetID,
		checkpointTableID,
	))

	// Execute the query.
	res, err := q.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to make read request to BigQuery: %w", err)
	}

	var rows []bigquery.Value
	nextErr := res.Next(&rows)

	// Table is empty, no cursor to provide
	if errors.Is(nextErr, iterator.Done) {
		return "", nil
	}

	if nextErr != nil {
		return "", fmt.Errorf("failed to iterate over query response: %w", nextErr)
	}

	checkpoint, ok := rows[0].(string)
	if !ok {
		return "", fmt.Errorf("error converting row value (%T) to int: %v", rows[0], ok)
	}

	return checkpoint, nil
}

// Write the latest checkpoint that was successfully processed.
// This is used by the retry service.
func (bq *BigQuery) WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt string) error {
	inserter := bq.client.Dataset(bq.datasetID).Table(checkpointTableID).Inserter()
	items := []*CheckpointEntry{
		// CheckpointEntry implements the ValueSaver interface.
		{deliveryID: deliveryID, createdAt: createdAt},
	}
	if err := inserter.Put(ctx, items); err != nil {
		return fmt.Errorf("failed to execute WriteCheckpointID: %w", err)
	}

	return nil
}

// Helper method to execute a count query and return the count.
func (bq *BigQuery) makeCountQuery(ctx context.Context, query string) (int, error) {
	// Construct a query.
	q := bq.client.Query(query)

	// Execute the query.
	res, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to make read request to BigQuery: %w", err)
	}

	var values []bigquery.Value
	if err := res.Next(&values); err != nil {
		return 0, fmt.Errorf("failed to iterate over query response: %w", err)
	}

	count, ok := values[0].(int)
	if !ok {
		return 0, fmt.Errorf("error converting row value (%T) to string: %w", values[0], err)
	}

	return count, nil
}

// Save implements the ValueSaver interface for a CheckpointEntry. A random insertID is generated by the library to facilitate deduplication.
func (ce *CheckpointEntry) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"delivery_id": ce.deliveryID,
		"created":     ce.createdAt,
	}, "", nil
}

// Save implements the ValueSaver interface for a FailureEventEntry. A random insertID is generated by the library to facilitate deduplication.
func (fe *FailureEventEntry) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"delivery_id": fe.deliveryID,
		"created":     fe.createdAt,
	}, "", nil
}
