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

// Package clients encapsulates all clients and associated helper methods to
// interact with other services.
package clients

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// BigQuery provides a client and dataset identifiers.
type BigQuery struct {
	projectID string
	datasetID string
	client    *bigquery.Client
	logger    *zap.SugaredLogger
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
		logger:    logging.FromContext(ctx).Named("bigquery"),
	}, nil
}

// Close releases any resources held by the BigQuery client.
func (bq *BigQuery) Close() error {
	if err := bq.client.Close(); err != nil {
		return fmt.Errorf("failed to close bigquery client: %w", err)
	}
	return nil
}

// Check if an entry with a given delivery_id already exists in the events
// table, this attempts to prevent duplicate processing of events. This is used
// by the webhook service.
func (bq *BigQuery) DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error) {
	res, err := bq.makeCountQuery(ctx, eventsTableID, deliveryID)
	if err != nil {
		return false, fmt.Errorf("failed to execute DeliveryEventExists: %w", err)
	}

	return res > 0, nil
}

// Check if the number of entries with a given delivery_id in the failure-events
// table exceeds the retry limit. This is used by the webhook service.
func (bq *BigQuery) FailureEventsExceedsRetryLimit(ctx context.Context, failureEventTableID, deliveryID string, maxRetry int) (bool, error) {
	res, err := bq.makeCountQuery(ctx, failureEventTableID, deliveryID)
	if err != nil {
		return false, fmt.Errorf("failed to execute FailureEventsExceedsRetryLimit: %w", err)
	}

	if res > 0 && res < maxRetry {
		bq.logger.Debugw("found retries, but does not exceed maxRetries",
			"retries", res,
			"delivery_id", deliveryID,
			"limit", maxRetry,
		)
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
		return fmt.Errorf("failed to execute WriteFailureEvent for deliveryID %s: %w", deliveryID, err)
	}

	return nil
}

// Retrieve the latest checkpoint cursor value (deliveryID) in the checkpoint
// table. This is used by the retry service.
func (bq *BigQuery) RetrieveCheckpointID(ctx context.Context, checkpointTableID string) (string, error) {
	// Construct a query.
	q := bq.client.Query("SELECT delivery_id FROM `@projectID.@datasetID.@checkpointTableID` ORDER BY created DESC LIMIT 1")

	// Set the params for the query
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "projectID",
			Value: bq.projectID,
		},
		{
			Name:  "datasetID",
			Value: bq.datasetID,
		},
		{
			Name:  "checkpointTableID",
			Value: checkpointTableID,
		},
	}

	// Execute the query.
	res, err := q.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to make read request to BigQuery: %w", err)
	}

	var rows []bigquery.Value
	nextErr := res.Next(&rows)

	if nextErr != nil {
		// if no checkpoint ID exists, return the empty string with no error
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
		return "", fmt.Errorf("failed to convert row value %v to string: (got %T)", rows[0], rows[0])
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
		return fmt.Errorf("failed to execute WriteCheckpointID for deliveryID %s: %w", deliveryID, err)
	}

	return nil
}

// Helper method to execute a count query for a given table by deliveryID and
// return the count.
func (bq *BigQuery) makeCountQuery(ctx context.Context, tableID, deliveryID string) (int, error) {
	q := bq.client.Query("SELECT COUNT(1) FROM `@projectID.@datasetID.@tableID` WHERE delivery_id = @deliveryID")

	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "projectID",
			Value: bq.projectID,
		},
		{
			Name:  "datasetID",
			Value: bq.datasetID,
		},
		{
			Name:  "tableID",
			Value: tableID,
		},
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

	count, ok := rows[0].(int)
	if !ok {
		return 0, fmt.Errorf("failed to convert row value %v to int (got %T)", rows[0], rows[0])
	}

	return count, nil
}

// Save implements the ValueSaver interface for a CheckpointEntry. A random
// insertID is generated by the library to facilitate deduplication.
func (ce *CheckpointEntry) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"delivery_id": ce.deliveryID,
		"created":     ce.createdAt,
	}, "", nil
}

// Save implements the ValueSaver interface for a FailureEventEntry. A random
// insertID is generated by the library to facilitate deduplication.
func (fe *FailureEventEntry) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"delivery_id": fe.deliveryID,
		"created":     fe.createdAt,
	}, "", nil
}
