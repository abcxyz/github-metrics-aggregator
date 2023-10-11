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

package review

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/github-metrics-aggregator/pkg/review/bq"
	"github.com/abcxyz/pkg/logging"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const (
	// GithubPRApproved is the value GitHub set the `reviewDecision` field to
	// when a Pull Request has been approved by a reviewer.
	GithubPRApproved = "APPROVED"

	// the default approval status we assign to a commit.
	DefaultApprovalStatus = "UNKNOWN"
)

// commitQuery is the BigQuery query that selects the commits that need
// to be processed. The criteria for a commit that needs to be processed are:
// 1. The commit was pushed to the repository's default branch.
// 2. We do not have a record for the commit in the commit_review_status table.
const commitQuery = `
SELECT
  push_events.pusher author,
  push_events.organization,
  push_events.repository,
  JSON_VALUE(commit_json, '$.id') commit_sha,
  JSON_VALUE(commit_json, '$.timestamp') commit_timestamp,
FROM
  ` + "`%s`" + ` push_events,
  UNNEST(push_events.commits) commit_json
LEFT JOIN
  ` + "`%s`" + ` commit_review_status
ON
  commit_review_status.commit_sha = commit_sha
WHERE
  push_events.ref = CONCAT('refs/heads/', push_events.repository_default_branch)
  AND commit_review_status.commit_sha IS NULL
`

// breakGlassIssueQuery is the BigQuery query that searches for a
// break glass issues created by given user and within a specified time frame.
const breakGlassIssueQuery = `
SELECT
  issues.html_url
FROM
  ` + "`%s`" + ` issues
WHERE
  issues.repository = 'breakglass'
  AND author = '%s'
  AND issues.created_at <= TIMESTAMP('%s')
  AND issues.closed_at >= TIMESTAMP('%s')
`

// Commit maps the columns from the driving BigQuery query
// to a usable structure.
type Commit struct {
	Author       string `bigquery:"author"`
	Organization string `bigquery:"organization"`
	Repository   string `bigquery:"repository"`
	Branch       string `bigquery:"branch"`
	SHA          string `bigquery:"commit_sha"`
	// Timestamp will be in ISO 8601 format (https://en.wikipedia.org/wiki/ISO_8601)
	// and should be parsable using time.RFC3339 format
	Timestamp string `bigquery:"commit_timestamp"`
}

// CommitReviewStatus maps the columns of the 'commit_review_status` table in
// BigQuery.
type CommitReviewStatus struct {
	Commit
	HTMLURL        string   `bigquery:"commit_html_url"`
	PullRequestID  int      `bigquery:"pull_request_id"`
	ApprovalStatus string   `bigquery:"approval_status"`
	BreakGlassURLs []string `bigquery:"break_glass_issue_urls"`
}

// breakGlassIssue is a struct that maps the columns of the result of
// the breakGlassIssueQuery.
type breakGlassIssue struct {
	HTMLURL string `biquery:"html_url"`
}

// PullRequest represents a pull request in GitHub and contains the
// GitHub assigned ID, the pull request number in the repository,
// and the review decision for the pull request.
type PullRequest struct {
	// BasRefName is the target the PR is being merged into. For example,
	// If a PR is being opened to merge the code from feature branch 'my-feature'
	// into branch 'main', then BasRefName for this PR would be 'main'.
	BaseRefName    githubv4.String
	DatabaseID     githubv4.Int
	Number         githubv4.Int
	ReviewDecision githubv4.String
}

// BreakGlassIssueFetcher fetches break glass issues from a data source.
type BreakGlassIssueFetcher interface {
	// getBreakGlassIssues retrieves all break glass issues created by the given
	// author and whose open duration contains the specified timestamp.
	// The issue's open duration contains the timestamp if
	// issue.created_at <= timestamp <= issue.closed_at holds.
	getBreakGlassIssues(ctx context.Context, author, timestamp string) ([]breakGlassIssue, error)
}

// BigQueryBreakGlassIssueFetcher implements the BreakGlassIssueFetcher
// interface and fetches the break glass issue data from BigQuery.
type BigQueryBreakGlassIssueFetcher struct {
	client *bigquery.Client
	config *CommitApprovalPipelineConfig
}

func (bqif *BigQueryBreakGlassIssueFetcher) getBreakGlassIssues(ctx context.Context, author, timestamp string) ([]breakGlassIssue, error) {
	issueQuery := GetBreakGlassIssueQuery(bqif.config.IssuesTable, author, timestamp)
	items, err := bq.Query[breakGlassIssue](ctx, bqif.client, issueQuery)
	if err != nil {
		return nil, fmt.Errorf("client.Query failed: %w", err)
	}
	return items, nil
}

// CommitApprovalPipelineConfig holds the configuration data for the
// commit approval beam pipeline.
type CommitApprovalPipelineConfig struct {
	// The token to use for authenticating with GitHub.
	GitHubAccessToken string
	// The fully qualified name of the BigQuery table that holds push event data.
	// This is the table that is used to source the commits that need to be
	// processed.
	PushEventsTable bigqueryio.QualifiedTableName
	// The fully qualified name of the BigQuery table that holds the commit
	// review/approval status. This is the table that stores the final output
	// of the pipeline.
	CommitReviewStatusTable bigqueryio.QualifiedTableName
	// The fully qualified name of the BigQuery table that holds GitHub issue
	// data. This table is used to determine if a commit was pushed using
	// 'break glass' permissions.
	IssuesTable bigqueryio.QualifiedTableName
}

// CommitApprovalDoFn is an object that implements beams "DoFn" interface to
// provide the processing logic for converting a Commit to CommitReviewStatus.
type CommitApprovalDoFn struct {
	Config       CommitApprovalPipelineConfig
	GithubClient *githubv4.Client
}

// BreakGlassIssueDoFn is an object that implements beams "DoFn" interface to
// provide the processing logic for converting retrieving the associated break
// glass issue for a CommitReviewStatus.
type BreakGlassIssueDoFn struct {
	Config                 CommitApprovalPipelineConfig
	BreakGlassIssueFetcher BreakGlassIssueFetcher
}

// ProcessElement is a DoFn implementation that take a Commit, determines
// if the commit was properly approved, and outputs the resulting
// CommitReviewStatus using the provided emit function.
// A commit is considered properly reviewed as long as there is an associated
// PR for the commit targeting the repository's main branch with reviewDecision
// of 'APPROVED'.
func (fn *CommitApprovalDoFn) ProcessElement(ctx context.Context, commit Commit, emit func(CommitReviewStatus)) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "processing commit", "commit", commit)
	requests, err := GetPullRequestsTargetingDefaultBranch(ctx, fn.GithubClient, commit.Organization, commit.Repository, commit.SHA)
	if err != nil {
		// There are essentially two different kind of errors that could happen:
		// 1. Transient Errors: We aren't able to get the pull requests for a commit
		//    because of some temporary issue with GitHub (e.g. GitHub servers are
		//    down). In these cases the commit will simply be retried the next time
		//    the pipeline is run.
		// 2. Permanent Errors: There is something wrong with the commit data itself
		//    that makes GitHub return an error. For example, the repository the
		//    commit came from may have been deleted so we no longer can get pull
		//    request information from GitHub about it.
		//
		// For the Transient Errors, the commit will be retried during the next
		// pipeline execution. So there is no need to do anything else aside from
		// logging the error.
		//
		// For Permanent Errors, it may be useful to do something aside from
		// logging, but it is hard to say exactly what should be done without seeing
		// what kinds of errors like this occur and how frequently. For now, we can
		// just to log the error and then consider more sophisticated error handling
		// if/when we need it.
		logger.ErrorContext(ctx, "failed to get pull requests for commit: %v", err)
		return // this commit could not be processed
	}
	commitReviewStatus := CommitReviewStatus{
		Commit:         commit,
		HTMLURL:        getCommitHTMLURL(commit),
		ApprovalStatus: DefaultApprovalStatus,
	}
	// GitHub's API is structured such that there may be more than one pull
	// request for a given commit in a repository. In practice this is very
	// unlikely to occur and there should only ever be one PR for each commit.
	// Regardless, we only care that there is at least one pull
	// request for the commit that has been approved by a reviewer. So we
	// will simply select the first PR we find that matches that criteria.
	pullRequest := getApprovingPullRequest(requests)
	// if there were no approving PRs, but we do have PRs for this commit, then
	// just choose the first one
	if pullRequest == nil && len(requests) > 0 {
		pullRequest = requests[0]
	}
	if pullRequest != nil {
		commitReviewStatus.PullRequestID = int(pullRequest.DatabaseID)
		commitReviewStatus.ApprovalStatus = string(pullRequest.ReviewDecision)
	}
	emit(commitReviewStatus)
}

// ProcessElement is a DoFn implementation that takes a CommitReviewStatus
// and populates its breakGlassIssue field (if necessary) and then emits it
// using the given emit function. The process only searches for break glass
// issues for commits that do not have the status GithubPRApproved.
func (fn *BreakGlassIssueDoFn) ProcessElement(ctx context.Context, commitReviewStatus CommitReviewStatus, emit func(status CommitReviewStatus)) {
	if emit == nil {
		panic("A nil emit function was passed in. Please check beam pipeline setup.")
	}
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "processing commitReviewStatus", "commitReviewStatus", commitReviewStatus)
	if commitReviewStatus.ApprovalStatus != GithubPRApproved {
		// if the commit does not have proper approval, we check if there was a
		// break glass issue opened by the author during the timeframe they
		// submitted the commit.
		breakGlassIssues, err := fn.BreakGlassIssueFetcher.getBreakGlassIssues(ctx, commitReviewStatus.Author, commitReviewStatus.Timestamp)
		if err != nil {
			// We should only get transient style errors from BigQuery
			// (e.g. network is down etc.). So, we can just log the error and then
			// drop this CommitReviewStatus from the pipeline. It will then get
			// retried on the next run of the pipeline.
			logger.ErrorContext(ctx, "failure when trying to get break glass issue: %v", err)
			return
		}
		commitReviewStatus.BreakGlassURLs = mapSlice(breakGlassIssues, func(issue breakGlassIssue) string {
			return issue.HTMLURL
		})
	}
	emit(commitReviewStatus)
}

// getApprovingPullRequest retrieves the first *PullRequest that has a
// review decision status with the value of GithubPRApproved. if no such
// *PullRequest is present then nil is returned.
func getApprovingPullRequest(pullRequests []*PullRequest) *PullRequest {
	for _, pullRequest := range pullRequests {
		if pullRequest.ReviewDecision == GithubPRApproved {
			return pullRequest
		}
	}
	return nil
}

func getCommitHTMLURL(commit Commit) string {
	return fmt.Sprintf("https://github.com/%s/%s/commit/%s", commit.Organization, commit.Repository, commit.SHA)
}

func NewGitHubGraphQLClient(ctx context.Context, accessToken string) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return githubv4.NewClient(httpClient)
}

// GetCommitQuery returns a BigQuery query that selects the commits that need
// to be processed.
func GetCommitQuery(pushEvents, commitReviewStatus bigqueryio.QualifiedTableName) string {
	return fmt.Sprintf(commitQuery, sqlFormat(pushEvents), sqlFormat(commitReviewStatus))
}

// GetBreakGlassIssueQuery returns a BigQuery query that searches for a
// break glass issue created by given user and within a specified time frame.
func GetBreakGlassIssueQuery(issues bigqueryio.QualifiedTableName, user, timestamp string) string {
	return fmt.Sprintf(breakGlassIssueQuery, sqlFormat(issues), user, timestamp, timestamp)
}

// sqlFormat formats the qualified name as "<project>.<dataset>.<table>"
// so that it can be used in SQL queries.
func sqlFormat(qualifiedTableName bigqueryio.QualifiedTableName) string {
	return fmt.Sprintf("%s.%s.%s", qualifiedTableName.Project, qualifiedTableName.Dataset, qualifiedTableName.Table)
}

// GetPullRequestsTargetingDefaultBranch retrieves all associated pull requests
// for a commit that target the repository's default branch from GitHub based on
// the given GitHub organization, repository, and commit sha. If the commit
// has no such associated pull requests then an empty slice is returned.
func GetPullRequestsTargetingDefaultBranch(ctx context.Context, client *githubv4.Client, githubOrg, repository, commitSha string) ([]*PullRequest, error) {
	var query struct {
		Repository struct {
			DefaultBranchRef struct {
				Name githubv4.String
			}
			Object struct {
				Commit struct {
					AssociatedPullRequest struct {
						Nodes    []*PullRequest
						PageInfo struct {
							EndCursor       githubv4.String
							HasNextPage     githubv4.Boolean
							HasPreviousPage githubv4.Boolean
							StartCursor     githubv4.String
						}
						TotalCount githubv4.Int
					} `graphql:"associatedPullRequests(first: 100, after: $pageCursor)"`
				} `graphql:"... on Commit"`
			} `graphql:"object(oid: $commitSha)"`
		} `graphql:"repository(owner: $githubOrg, name: $repository)"`
	}
	getPage := func(cursor githubv4.String) error {
		err := client.Query(ctx, &query, map[string]interface{}{
			"githubOrg":  githubv4.String(githubOrg),
			"repository": githubv4.String(repository),
			"commitSha":  githubv4.GitObjectID(commitSha),
			"pageCursor": cursor,
		})
		if err != nil {
			return fmt.Errorf("GitHub GraphgQL call failed: %w", err)
		}
		return nil
	}
	if err := getPage(""); err != nil {
		return nil, err
	}
	pullRequests := make([]*PullRequest, 0, query.Repository.Object.Commit.AssociatedPullRequest.TotalCount)
	pullRequests = append(pullRequests, query.Repository.Object.Commit.AssociatedPullRequest.Nodes...)
	for query.Repository.Object.Commit.AssociatedPullRequest.PageInfo.HasNextPage {
		pageCursor := query.Repository.Object.Commit.AssociatedPullRequest.PageInfo.EndCursor
		if err := getPage(pageCursor); err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, query.Repository.Object.Commit.AssociatedPullRequest.Nodes...)
	}
	pullRequests = filter(pullRequests, func(pullRequest *PullRequest) bool {
		return pullRequest.BaseRefName == query.Repository.DefaultBranchRef.Name
	})
	return pullRequests, nil
}

func filter[T any](slice []T, predicate func(T) bool) []T {
	filtered := make([]T, 0)
	for _, t := range slice {
		if predicate(t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func mapSlice[T, U any](slice []T, mapper func(T) U) []U {
	mapped := make([]U, 0, len(slice))
	for _, t := range slice {
		mapped = append(mapped, mapper(t))
	}
	return mapped
}
