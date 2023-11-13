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

// Package review contains code to get review status information
// for a GitHub commit.

package comment

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/github-metrics-aggregator/pkg/comment/bq"
	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

//go:embed sql/prWorkflowsToCommentQuery.sql
var prWorkflowsToCommentQuery string

//go:embed sql/find_start_time.sql
var findStartTimeQuery string

// InvocationCommentStatus maps the columns of the 'invocation_comment_status` table in
// BigQuery.
type InvocationCommentStatus struct {
	PullRequestID      int64              `bigquery:"pull_request_id"`
	PullRequestHTMLURL string             `bigquery:"pull_request_html_url"`
	ProcessedAt        time.Time          `bigquery:"processed_at"`
	CommentID          bigquery.NullInt64 `bigquery:"comment_id"`
	Status             string             `bigquery:"status"`
	JobName            string             `bigquery:"job_name"`
}

type PrCommentInfo struct {
	Org      string
	Repo     string
	PrNumber int
}

type WorkflowToProcess struct {
	WorkflowDeliveryID string    `bigquery:"workflow_delivery_id"`
	WorkflowURI        string    `bigquery:"workflow_uri"`
	LogsURI            string    `bigquery:"logs_uri"`
	PrID               int64     `bigquery:"pr_id"`
	PrURL              string    `bigquery:"pr_url"`
	WorkflowPayload    string    `bigquery:"workflow_payload"`
	PrReceivedTime     time.Time `bigquery:"pr_received_time"`
}

// PrCommentPipelineConfig holds the configuration data for the
// pr comment beam pipeline.
type PrCommentPipelineConfig struct {
	// The token to use for authenticating with GitHub.
	GitHubAccessToken string
	// GCP project id that holds GitHub Metrics Aggregator tables.
	GcpProjectID string
	// BigQuery dataset ID that holds GitHub Metrics Aggregator tables.
	BigQueryDatasetID string
	// Only PRs merged on or after this time will be considered for processing.
	StartProcessingTime time.Time
}

type StartTimeResult struct {
	Time time.Time `bigquery:"start_time"`
}

func ProcessABatchOfRrsAndWorkflows(ctx context.Context, config PrCommentPipelineConfig) error {
	logger := logging.FromContext(ctx)
	logger.Debug("creating clients")
	ghClient := NewGitHubClient(ctx, config.GitHubAccessToken)
	bqClient, err := NewBigQueryClient(ctx, config)
	if err != nil {
		return err
	}
	defer bqClient.Close() // TODO: handle err?
	logger.Debug("getting start time")
	startTime, err := GetStartTime(ctx, bqClient, config)
	if err != nil {
		return err
	}
	logger.Debug("getting workflows to process")
	workflowsToProcess, err := GetWorkflowsToProcess(ctx, bqClient, config, startTime)
	if err != nil {
		return err
	}
	if len(workflowsToProcess) == 0 {
		logger.Info("no workflows to process for now, finishing up")
		return nil
	}

}

func CommentOnGitHub(ctx context.Context, ghClient *github.Client) {
	createdComment, response, err := ghClient.Issues.CreateComment(ctx, org, repo, prNumber, comment)
}

// GetStartTime finds a stricter bound on the start time to reduce query costs.
// Because PRs are processed in order, we just need
// MAX(configuredStart, MIN(earliest_retryable_failure, latest_successful_pr).
func GetStartTime(ctx context.Context, bqClient *bigquery.Client, config PrCommentPipelineConfig) (time.Time, error) {
	logger := logging.FromContext(ctx)
	params := []bigquery.QueryParameter{{Name: "earliest_timestamp", Value: config.StartProcessingTime}}
	startTime, err := bq.Query[StartTimeResult](ctx, bqClient, findStartTimeQuery, config.GcpProjectID, config.BigQueryDatasetID, params)
	if err != nil {
		logger.Warn("error when fetching stricter start time bound from BigQuery, will use coarse lower bound instead", "error", err)
		return config.StartProcessingTime, nil
	}
	if len(startTime) != 1 {
		// You could be more defensive and fallback to default upper bound, but this should never
		// happen, so I want to know if it does!
		return time.Time{}, fmt.Errorf("unexpected number of results when reading stricter start time bound from BigQuery: %v", len(startTime))
	}
	return startTime[0].Time, nil
}

func GetWorkflowsToProcess(ctx context.Context, bqClient *bigquery.Client, config PrCommentPipelineConfig, startTime time.Time) ([]WorkflowToProcess, error) {
	// Don't look at PRs merged too recently, we don't want to miss any post-merge workflows.
	prEndTime := time.Now().UTC().Add(-1 * time.Hour)
	// We will look for workflows up to a month earlier than our pr merge start.
	workflowStartTime := startTime.Add(-1 * time.Hour)
	// We want all possible workflows associated with any PRs, so go to current time.
	workflowEndTime := time.Now().UTC()
	params := []bigquery.QueryParameter{
		{Name: "ts_pr_search_start", Value: startTime},
		{Name: "ts_pr_search_end", Value: prEndTime},
		{Name: "ts_workflow_search_start", Value: workflowStartTime},
		{Name: "ts_workflow_search_end", Value: workflowEndTime},
	}
	workflows, err := bq.Query[WorkflowToProcess](ctx, bqClient, prWorkflowsToCommentQuery, config.GcpProjectID, config.BigQueryDatasetID, params)
	if err != nil {
		return nil, fmt.Errorf("error when fetching workflows to process from BigQuery: %w", err)
	}
	return workflows, nil
}

func NewBigQueryClient(ctx context.Context, config PrCommentPipelineConfig) (*bigquery.Client, error) {
	// Should use application default credentials.
	client, err := bigquery.NewClient(ctx, config.GcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("unable to create bigquery client: %w", err)
	}
	return client, nil
}

func NewGitHubClient(ctx context.Context, accessToken string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return github.NewClient(httpClient)
}
