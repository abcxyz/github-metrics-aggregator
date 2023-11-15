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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/abcxyz/github-metrics-aggregator/pkg/review/bq"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// init registers the DoFns used in this pipeline with apache beam.
// This allows the beam SDK to infer an encoding from any inputs/outputs,
// registers the DoFn for execution on remote runners, and optimizes the
// runtime execution of the DoFns via reflection.
func init() {
	beam.RegisterType(reflect.TypeOf((*CommitApprovalDoFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*BreakGlassIssueDoFn)(nil)))
	register.DoFn3x0[context.Context, Commit, func(status CommitReviewStatus)](&CommitApprovalDoFn{})
	register.DoFn3x0[context.Context, CommitReviewStatus, func(status CommitReviewStatus)](&BreakGlassIssueDoFn{})
	register.Emitter1[CommitReviewStatus]()
}

const (
	// GithubPRApproved is the value GitHub set the `reviewDecision` field to
	// when a Pull Request has been approved by a reviewer.
	GithubPRApproved = "APPROVED"

	// the default approval status we assign to a commit.
	DefaultApprovalStatus = "UNKNOWN"
)

// commitSQL is the BigQuery query that selects the commits that need
// to be processed. The criteria for a commit that needs to be processed are:
// 1. The commit was pushed to the repository's default branch.
// 2. We do not have a record for the commit in the commit_review_status table.
const commitSQL = `
SELECT
  push_events.pusher author,
  push_events.organization,
  push_events.repository,
  push_events.repository_default_branch branch,
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

// breakGlassIssueSQL is the BigQuery query that searches for a
// break glass issues created by given user and within a specified time frame.
const breakGlassIssueSQL = `
SELECT
  issues.html_url html_url
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
	HTMLURL            string   `bigquery:"commit_html_url"`
	PullRequestID      int      `bigquery:"pull_request_id"`
	PullRequestNumber  int      `bigquery:"pull_request_number"`
	PullRequestHTMLURL string   `bigquery:"pull_request_html_url"`
	ApprovalStatus     string   `bigquery:"approval_status"`
	BreakGlassURLs     []string `bigquery:"break_glass_issue_urls"`
	Note               string   `bigquery:"note"`
}

// breakGlassIssue is a struct that maps the columns of the result of
// the breakGlassIssueQuery.
type breakGlassIssue struct {
	HTMLURL string `bigquery:"html_url"`
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
	URL            githubv4.String
}

// BreakGlassIssueFetcher fetches break glass issues from a data source.
type BreakGlassIssueFetcher interface {
	// getBreakGlassIssues retrieves all break glass issues created by the given
	// author and whose open duration contains the specified timestamp.
	// The issue's open duration contains the timestamp if
	// issue.created_at <= timestamp <= issue.closed_at holds.
	getBreakGlassIssues(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error)
}

// BigQueryBreakGlassIssueFetcher implements the BreakGlassIssueFetcher
// interface and fetches the break glass issue data from BigQuery.
type BigQueryBreakGlassIssueFetcher struct {
	client *bigquery.Client
	config *CommitApprovalPipelineConfig
}

func (bqif *BigQueryBreakGlassIssueFetcher) getBreakGlassIssues(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error) {
	issueQuery := breakGlassIssueQuery(bqif.config.IssuesTable, author, timestamp)
	items, err := bq.Query[breakGlassIssue](ctx, bqif.client, issueQuery)
	if err != nil {
		return nil, fmt.Errorf("client.Query failed: %w", err)
	}
	return items, nil
}

// CommitApprovalPipelineConfig holds the configuration data for the
// commit approval beam pipeline.
type CommitApprovalPipelineConfig struct {
	// GitHubToken is the GitHub Token to use for authentication
	GitHubToken string

	// PushEventsTable is the fully qualified name of the BigQuery table that
	// holds push event data. This is the table that is used to source the commits
	// that need to be processed.
	PushEventsTable *bigqueryio.QualifiedTableName

	// CommitReviewStatusTable is the fully qualified name of the BigQuery table
	// that holds the commit review/approval status. This is the table that stores
	// the final output of the pipeline.
	CommitReviewStatusTable *bigqueryio.QualifiedTableName

	// IssuesTable is the fully qualified name of the BigQuery table that holds GitHub issue
	// data. This table is used to determine if a commit was pushed using
	// 'break glass' permissions.
	IssuesTable *bigqueryio.QualifiedTableName
}

// CommitApprovalDoFn is an object that implements beams "DoFn" interface to
// provide the processing logic for converting a Commit to CommitReviewStatus.
type CommitApprovalDoFn struct {
	// Config is the configuration. Beam will serialize public attributes of the
	// struct when intitializing worker nodes. Thus any attribute that should be
	// serialized needs to be public.
	Config *CommitApprovalPipelineConfig

	// The following attributes do not properly support serialization. Thus,
	// we will make them private to avoid Beam from trying to serialize them.
	// Instead, they will be lazy initialized during pipeline execution when
	// StartBundle is called.
	githubClient *githubv4.Client
}

// BreakGlassIssueDoFn is an object that implements beams "DoFn" interface to
// provide the processing logic for converting retrieving the associated break
// glass issue for a CommitReviewStatus.
type BreakGlassIssueDoFn struct {
	// Config is the configuration. Beam will serialize public attributes of the
	// struct when intitializing worker nodes. Thus any attribute that should be
	// serialized needs to be public.
	Config *CommitApprovalPipelineConfig

	// The following attributes do not properly support serialization. Thus,
	// we will make them private to avoid Beam from trying to serialize them.
	// Instead, they will be lazy initialized during pipeline execution when
	// StartBundle is called.
	breakGlassIssueFetcher BreakGlassIssueFetcher
}

// NewCommitApprovalPipeline constructs and returns a *beam.Pipeline to get
// approval status for commits.
func NewCommitApprovalPipeline(config *CommitApprovalPipelineConfig) *beam.Pipeline {
	pipeline, scope := beam.NewPipelineWithRoot()

	// Step 1: Get commits that need to be processed from BigQuery.
	query := commitQuery(config.PushEventsTable, config.CommitReviewStatusTable)
	commits := bigqueryio.Query(scope, config.PushEventsTable.Project, query, reflect.TypeOf(Commit{}), bigqueryio.UseStandardSQL())

	// Step 2: Get review status information for each commit.
	reviewStatuses := beam.ParDo(scope, &CommitApprovalDoFn{Config: config}, commits)

	// Step 3: Look up break glass issue if necessary.
	taggedReviewStatuses := beam.ParDo(scope, &BreakGlassIssueDoFn{Config: config}, reviewStatuses)

	// Step 4: Write the commit review status information to BigQuery.
	bigqueryio.Write(scope, config.CommitReviewStatusTable.Project, config.CommitReviewStatusTable.String(), taggedReviewStatuses)
	return pipeline
}

// StartBundle is called by Beam when the DoFn function is initialized. With a
// local runner this is called from the running version of the application. For
// Dataflow, this is called on each worker node after the binary is provisioned.
// Remote Dataflow workers do not have the same environment or runtime arguments
// as the launcher process. The CommitApprovalDoFn struct is serialized to the
// worker along with all public attributes that can be serialized. This causes
// us to have to initialize the githubClient from this method once it has
// materialized on the remote host. Since ProcessElement uses an emit function,
// we are required by Beam to accept one in StartBundle as well even though it
// is not used.
func (fn *CommitApprovalDoFn) StartBundle(ctx context.Context, _ func(CommitReviewStatus)) error {
	fn.githubClient = NewGitHubGraphQLClient(ctx, fn.Config.GitHubToken)
	return nil
}

// ProcessElement is a DoFn implementation that take a Commit, determines
// if the commit was properly approved, and outputs the resulting
// CommitReviewStatus using the provided emit function.
// A commit is considered properly reviewed as long as there is an associated
// PR for the commit targeting the repository's main branch with reviewDecision
// of 'APPROVED'.
func (fn *CommitApprovalDoFn) ProcessElement(ctx context.Context, commit Commit, emit func(CommitReviewStatus)) {
	// beam/log is required in order for log severity to show up properly in
	// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
	// for more context.
	log.Infof(ctx, "processing commit: %+v", commit)
	commitReviewStatus := CommitReviewStatus{
		Commit:         commit,
		HTMLURL:        getCommitHTMLURL(commit),
		ApprovalStatus: DefaultApprovalStatus,
		BreakGlassURLs: make([]string, 0),
	}
	requests, err := GetPullRequestsTargetingDefaultBranch(ctx, fn.githubClient, commit.Organization, commit.Repository, commit.SHA)
	if err != nil {
		// Special error cases
		if strings.HasPrefix(err.Error(), "failed to call graphql") {
			unwrapped := errors.Unwrap(err)
			if strings.HasPrefix(unwrapped.Error(), "Could not resolve to a Repository") {
				// this is a permanent error from GitHub telling us the repository
				// for the commit no longer exists. Note this in the commit review status
				// and send it on for further processing
				commitReviewStatus.Note = unwrapped.Error()
				emit(commitReviewStatus)
				return // finished with this commit
			}
		}
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
		//
		// beam/log is required in order for log severity to show up properly in
		// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
		// for more context.
		log.Errorf(ctx, "failed to get pull requests for commit: %v", err)
		return // this commit could not be processed
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
		commitReviewStatus.PullRequestNumber = int(pullRequest.Number)
		commitReviewStatus.PullRequestHTMLURL = string(pullRequest.URL)
		commitReviewStatus.ApprovalStatus = string(pullRequest.ReviewDecision)
	}
	emit(commitReviewStatus)
}

// StartBundle is called by Beam when the DoFn function is initialized. With a
// local runner this is called from the running version of the application. For
// Dataflow, this is called on each worker node after the binary is provisioned.
// Remote Dataflow workers do not have the same environment or runtime arguments
// as the launcher process. The CommitApprovalDoFn struct is serialized to the
// worker along with all public attributes that can be serialized. This causes
// us to have to initialize the bigQueryClient from this method once it has
// materialized on the remote host. Since ProcessElement uses an emit function,
// we are required by Beam to accept one in StartBundle as well even though it
// is not used.
func (fn *BreakGlassIssueDoFn) StartBundle(ctx context.Context, _ func(CommitReviewStatus)) error {
	// initialize break glass issue fetcher
	bigQueryClient, err := bigquery.NewClient(ctx, fn.Config.IssuesTable.Project)
	if err != nil {
		return fmt.Errorf("failed to construct bigquery client: %w", err)
	}
	fn.breakGlassIssueFetcher = &BigQueryBreakGlassIssueFetcher{
		client: bigQueryClient,
		config: fn.Config,
	}
	return nil
}

// ProcessElement is a DoFn implementation that takes a CommitReviewStatus
// and populates its breakGlassIssue field (if necessary) and then emits it
// using the given emit function. The process only searches for break glass
// issues for commits that do not have the status GithubPRApproved.
func (fn *BreakGlassIssueDoFn) ProcessElement(ctx context.Context, commitReviewStatus CommitReviewStatus, emit func(status CommitReviewStatus)) {
	if emit == nil {
		panic("A nil emit function was passed in. Please check beam pipeline setup.")
	}
	// beam/log is required in order for log severity to show up properly in
	// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
	// for more context.
	log.Infof(ctx, "processing commitReviewStatus: %+v", commitReviewStatus)
	if commitReviewStatus.ApprovalStatus != GithubPRApproved {
		// if the commit does not have proper approval, we check if there was a
		// break glass issue opened by the author during the timeframe they
		// submitted the commit.
		breakGlassIssues, err := fn.breakGlassIssueFetcher.getBreakGlassIssues(ctx, commitReviewStatus.Author, commitReviewStatus.Timestamp)
		if err != nil {
			// We should only get transient style errors from BigQuery
			// (e.g. network is down etc.). So, we can just log the error and then
			// drop this CommitReviewStatus from the pipeline. It will then get
			// retried on the next run of the pipeline.
			//
			// beam/log is required in order for log severity to show up properly in
			// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
			// for more context.
			log.Errorf(ctx, "failure when trying to get break glass issue: %v", err)
			return
		}

		for _, v := range breakGlassIssues {
			commitReviewStatus.BreakGlassURLs = append(commitReviewStatus.BreakGlassURLs, v.HTMLURL)
		}
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

// commitQuery returns a BigQuery query that selects the commits that need to be
// processed.
func commitQuery(pushEvents, commitReviewStatus *bigqueryio.QualifiedTableName) string {
	return fmt.Sprintf(commitSQL, formatGoogleSQL(pushEvents), formatGoogleSQL(commitReviewStatus))
}

// breakGlassIssueQuery returns a BigQuery query that searches for a break glass
// issue created by given user and within a specified time frame.
func breakGlassIssueQuery(issues *bigqueryio.QualifiedTableName, user, timestamp string) string {
	return fmt.Sprintf(breakGlassIssueSQL, formatGoogleSQL(issues), user, timestamp, timestamp)
}

// formatGoogleSQL formats the qualified table name in GoogleSQL syntax.
// i.e. "<project>.<dataset>.<table>".
func formatGoogleSQL(qualifiedTableName *bigqueryio.QualifiedTableName) string {
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

	pullRequests := make([]*PullRequest, 0, query.Repository.Object.Commit.AssociatedPullRequest.TotalCount)
	cursor := githubv4.String("")
	for {
		if err := client.Query(ctx, &query, map[string]any{
			"githubOrg":  githubv4.String(githubOrg),
			"repository": githubv4.String(repository),
			"commitSha":  githubv4.GitObjectID(commitSha),
			"pageCursor": cursor,
		}); err != nil {
			return nil, fmt.Errorf("failed to call graphql: %w", err)
		}

		// Only select pull requests made against the default branch.
		for _, pr := range query.Repository.Object.Commit.AssociatedPullRequest.Nodes {
			if pr.BaseRefName == query.Repository.DefaultBranchRef.Name {
				pullRequests = append(pullRequests, pr)
			}
		}

		if !query.Repository.Object.Commit.AssociatedPullRequest.PageInfo.HasNextPage {
			break
		}
		cursor = query.Repository.Object.Commit.AssociatedPullRequest.PageInfo.EndCursor
	}
	return pullRequests, nil
}
