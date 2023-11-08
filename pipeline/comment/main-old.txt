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

// Package main contains a Beam data pipeline that will read workflow event
// from BigQuery, comment on PRs with links to logs stored by leech, and
// store status back to BigQuery.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/github-metrics-aggregator/pkg/comment"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	githubToken                  string
	uniqueEventsTableFunction    string
	invocationCommentStatusTable string
	leechTable                   string
)

// TODO: move this struct to comment once lock14's change to use a github app is finished
// InvocationCommentStatus maps the columns of the 'invocation_comment_status` table in
// BigQuery.
type InvocationCommentStatus struct {
	PullRequestID      int64              `bigquery:"pull_request_id"`
	PullRequestHTMLURL string             `bigquery:"pull_request_html_url"`
	ProcessedAt        time.Time          `bigquery:"processed_at"`
	CommentId          bigquery.NullInt64 `bigquery:"comment_id"`
	Status             string             `bigquery:"status"`
	JobName            string             `bigquery:"job_name"`
	RetryJobAttempts   bigquery.NullInt64 `bigquery:"retry_job_attempts"`
}

func init() {
	// setup commandline arguments
	// explicitly *not* using the cli interface from abcxyz/pkg/cli due to conflicts
	// with Beam while using the Dataflow runner.
	flag.StringVar(&githubToken, "github-token", "", "The token to use to authenticate with github.")
	flag.StringVar(&uniqueEventsTableFunction, "unique-events-table", "", "The name of the unique events table function. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&invocationCommentStatusTable, "invocation-comment-status-table", "", "The name of the pr comment status table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&leechTable, "leech-table", "", "The name of the leech table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		// beam/log is required in order for log severity to show up properly in
		// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
		// for more context.
		log.Errorf(ctx, "realMain failed: %v", err)
		os.Exit(1)
	}
}

// realMain executes the PR Comment Pipeline.
func realMain(ctx context.Context) error {
	// parse flags
	flag.Parse()
	// initialize beam. This must be done after flag.Parse() but before any flag validation.
	// https://cloud.google.com/dataflow/docs/guides/setting-pipeline-options#CreatePipelineFromArgs
	beam.Init()
	// parse commandline arguments into config
	pipelineConfig, err := getConfigFromFlags()
	if err != nil {
		return fmt.Errorf("failed to construct pipeline config: %w", err)
	}
	// construct and execute the pipeline
	pipeline := comment.NewCommitApprovalPipeline(pipelineConfig)
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

// getConfigFromFlags returns comment.CommitApprovalPipelineConfig struct
// using flag values. returns an error if any of the flags have malformed
// data.
func getConfigFromFlags() (*comment.CommitApprovalPipelineConfig, error) {
	if githubToken == "" {
		return nil, fmt.Errorf("a non-empty github-token must be provided")
	}
	qualifiedPushEventsTable, err := bigqueryio.NewQualifiedTableName(pushEventsTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pushEventsTable: %w", err)
	}
	qualifiedCommitReviewStatusTable, err := bigqueryio.NewQualifiedTableName(invocationCommentStatusTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse invocationCommentStatusTable: %w", err)
	}
	qualifiedIssuesTable, err := bigqueryio.NewQualifiedTableName(issuesTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse issuesTable: %w", err)
	}
	return &comment.CommitApprovalPipelineConfig{
		GitHubAccessToken:       githubToken,
		PushEventsTable:         qualifiedPushEventsTable,
		CommitReviewStatusTable: qualifiedCommitReviewStatusTable,
		IssuesTable:             qualifiedIssuesTable,
	}, nil
}
