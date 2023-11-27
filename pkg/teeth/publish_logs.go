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

// Package teeth contains a job that will read leech and pull request event
// records from BigQuery and publish any available log invocations in a PR
// comment.
package teeth

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"
	"time"

	"cloud.google.com/go/bigquery"
)

// PublisherSourceQuery is the source query that teeth job pipeline
// will use to publish results.
//
//go:embed sql/publisher_source.sql
var PublisherSourceQuery string

// BigQueryClient defines the spec for calls to read from and write to
// BigQuery tables.
type BigQueryClient interface {
	Config() *BQConfig
	Query(string) *bigquery.Query
}

// TODO: Add query limit param.
//
// BQConfig defines configuration parameters for the BigQuery tables
// used by the teeth job pipeline.
type BQConfig struct {
	PullRequestEventsTable       string
	InvocationCommentStatusTable string
	EventsTable                  string
	LeechStatusTable             string
}

// PublisherSourceRecord maps the columns from the driving BigQuery query
// to a usable structure.
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

func SetUpPublisherSourceQuery(ctx context.Context, bqClient BigQueryClient) (*bigquery.Query, error) {
	tmpl, err := template.New("publisher").Parse(PublisherSourceQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to set up sql template: %w", err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, bqClient.Config()); err != nil {
		return nil, fmt.Errorf("failed to execute sql template: %w", err)
	}
	return bqClient.Query(b.String()), nil
}
