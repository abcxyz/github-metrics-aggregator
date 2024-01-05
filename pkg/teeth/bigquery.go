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

package teeth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	_ "embed"
)

// PublisherSourceQuery is the source query that teeth job pipeline
// will use to publish results.
//
//go:embed sql/publisher_source.sql
var PublisherSourceQuery string

// TODO: Add query limit param.
//
// BQConfig defines configuration parameters for the BigQuery client
// and the tables used by the teeth job pipeline.
type BQConfig struct {
	ProjectID string
	DatasetID string

	PullRequestEventsTable       string
	InvocationCommentStatusTable string
	EventsTable                  string
	LeechStatusTable             string
}

// PublisherSourceRecord maps the columns from the source query
// to a struct.
type PublisherSourceRecord struct {
	DeliveryID     string    `bigquery:"delivery_id"`
	PullRequestID  int       `bigquery:"pull_request_id"`
	PullRequestURL string    `bigquery:"pull_request_html_url"`
	Received       time.Time `bigquery:"received"`
	LogsURI        string    `bigquery:"logs_uri"`
	HeadSHA        string    `bigquery:"head_sha"`
}

// InvocationCommentStatusRecord is the output data structure that maps to the
// teeth pipeline's output table schema for invocation comment statuses.
type InvocationCommentStatusRecord struct {
	PullRequestID  int                `bigquery:"pull_request_id"`
	PullRequestURL string             `bigquery:"pull_request_html_url"`
	ProcessedAt    time.Time          `bigquery:"processed_at"`
	CommentID      bigquery.NullInt64 `bigquery:"comment_id"`
	Status         string             `bigquery:"status"`
	JobName        string             `bigquery:"job_name"`
}

// BigQuery provides a client to BigQuery API.
type BigQuery struct {
	config *BQConfig
	client *bigquery.Client
}

// NewBigQuery creates a new instance of a BigQuery client with config.
func NewBigQuery(ctx context.Context, config *BQConfig) (*BigQuery, error) {
	c, err := bigquery.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery.Client: %w", err)
	}
	return &BigQuery{
		config: config,
		client: c,
	}, nil
}

// Close closes the BigQuery client.
func (bq *BigQuery) Close() error {
	return bq.client.Close()
}

// Config returns the BigQuery client config
func (bq *BigQuery) Config() *BQConfig {
	return bq.config
}

// ExecutePublisherSourceQuery takes a Query implementation of the
// PublisherSourceQuery and runs it on BigQuery.
//
// This is normally called after calling SetUpPublisherSourceQuery.
//
// Returns the PublisherSourceQuery results.
func (bq *BigQuery) QueryLatest(ctx context.Context) ([]*PublisherSourceRecord, error) {
	qStr, err := populatePublisherSourceQuery(ctx, bq.config)
	if err != nil {
		return nil, fmt.Errorf("failed to populate query: %w", err)
	}

	// below copied from https://pkg.go.dev/cloud.google.com/go/bigquery#hdr-Querying
	q := bq.client.Query(qStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}
	results := make([]*PublisherSourceRecord, 0)
	for {
		r := &PublisherSourceRecord{}
		err := it.Next(r)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read result: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}

func populatePublisherSourceQuery(ctx context.Context, config *BQConfig) (string, error) {
	type Tables struct {
		PullRequestEventsTable       string
		InvocationCommentStatusTable string
		EventsTable                  string
		LeechStatusTable             string
	}
	tablePrefix := fmt.Sprintf("%s.%s.", config.ProjectID, config.DatasetID)
	fullTableNames := &Tables{
		PullRequestEventsTable:       tablePrefix + config.PullRequestEventsTable,
		InvocationCommentStatusTable: tablePrefix + config.InvocationCommentStatusTable,
		EventsTable:                  tablePrefix + config.EventsTable,
		LeechStatusTable:             tablePrefix + config.LeechStatusTable,
	}
	tmpl, err := template.New("publisher").Parse(PublisherSourceQuery)
	if err != nil {
		return "", fmt.Errorf("failed to set up sql template: %w", err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, fullTableNames); err != nil {
		return "", fmt.Errorf("failed to execute sql template: %w", err)
	}
	return b.String(), nil
}

// Insert writes statuses to the InvocationCommentStatusTable.
func (bq *BigQuery) Insert(ctx context.Context, statuses []*InvocationCommentStatusRecord) error {
	datasetID := bq.config.DatasetID
	tableID := bq.config.InvocationCommentStatusTable
	inserter := bq.client.Dataset(datasetID).Table(tableID).Inserter()
	if err := inserter.Put(ctx, statuses); err != nil {
		return fmt.Errorf("failed to insert statuses: %w", err)
	}
	return nil
}
