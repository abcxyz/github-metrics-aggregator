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

// Package main contains a Beam data pipeline that will read commit data
// from BigQuery, check each commits approval status, and write the approval
// status back to BigQuery.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/review"
	"github.com/abcxyz/pkg/logging"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	githubToken             string
	pushEventsTable         string
	commitReviewStatusTable string
	issuesTable             string
)

func init() {
	// setup commandline arguments
	// explicitly *not* using the cli interface from abcxyz/pkg/cli due to conflicts
	// with Beam while using the Dataflow runner.
	flag.StringVar(&githubToken, "github-token", "", "The token to use to authenticate with github.")
	flag.StringVar(&pushEventsTable, "push-events-table", "", "The name of the push events table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&commitReviewStatusTable, "commit-review-status-table", "", "The of the commit review status table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&issuesTable, "issues-table", "", "The name of the issues table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	logger := logging.FromContext(ctx)

	if err := realMain(ctx); err != nil {
		done()
		logger.ErrorContext(ctx, "realMain failed", err)
		os.Exit(1)
	}
}

// realMain executes the Commit Review Pipeline.
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
	pipeline := review.NewCommitApprovalPipeline(pipelineConfig)
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

// getConfigFromFlags returns review.CommitApprovalPipelineConfig struct
// using flag values. returns an error if any of the flags have malformed
// data.
func getConfigFromFlags() (*review.CommitApprovalPipelineConfig, error) {
	if githubToken == "" {
		return nil, fmt.Errorf("a non-empty github-token must be provided")
	}
	qualifiedPushEventsTable, err := bigqueryio.NewQualifiedTableName(pushEventsTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pushEventsTable: %w", err)
	}
	qualifiedCommitReviewStatusTable, err := bigqueryio.NewQualifiedTableName(commitReviewStatusTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse commitReviewStatusTable: %w", err)
	}
	qualifiedIssuesTable, err := bigqueryio.NewQualifiedTableName(issuesTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse issuesTable: %w", err)
	}
	return &review.CommitApprovalPipelineConfig{
		GitHubAccessToken:       githubToken,
		PushEventsTable:         qualifiedPushEventsTable,
		CommitReviewStatusTable: qualifiedCommitReviewStatusTable,
		IssuesTable:             qualifiedIssuesTable,
	}, nil
}
