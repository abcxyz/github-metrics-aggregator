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

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// githubPRApproved is the value GitHub set the `reviewDecision` field to
// when a Pull Request has been approved by a reviewer.
const githubPRApproved = "APPROVED"

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
	SHA          string `bigquery:"commit_sha"`
	// Timestamp will be in ISO 8601 format (https://en.wikipedia.org/wiki/ISO_8601)
	// and should be parsable using time.RFC3339 format
	Timestamp string `bigquery:"commit_timestamp"`
}

// PullRequest represents a pull request in GitHub and contains the
// GitHub assigned ID, the pull request number in the repository,
// and the review decision for the pull request.
type PullRequest struct {
	DatabaseID     githubv4.Int
	Number         githubv4.Int
	ReviewDecision githubv4.String
}

// getApprovingPullRequest retrieves the first *PullRequest that has a
// review decision status with the value of githubPRApproved. if no such
// *PullRequest is present then nil is returned.
func getApprovingPullRequest(pullRequests []*PullRequest) *PullRequest {
	for _, pullRequest := range pullRequests {
		if pullRequest.ReviewDecision == githubPRApproved {
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
func GetCommitQuery(project, dataset, pushEventsTable, commitReviewStatusTable string) string {
	pushEvents := fmt.Sprintf("%s.%s.%s", project, dataset, pushEventsTable)
	commitReviewStatus := fmt.Sprintf("%s.%s.%s", project, dataset, commitReviewStatusTable)
	return fmt.Sprintf(commitQuery, pushEvents, commitReviewStatus)
}

// GetBreakGlassIssueQuery returns a BigQuery query that searches for a
// break glass issue created by given user and within a specified time frame.
func GetBreakGlassIssueQuery(project, dataset, issuesTable, user, timestamp string) string {
	issues := fmt.Sprintf("%s.%s.%s", project, dataset, issuesTable)
	return fmt.Sprintf(breakGlassIssueQuery, issues, user, timestamp, timestamp)
}

// GetPullRequests retrieves all associated pull requests for a commit from GitHub based on
// the given GitHub organization, repository, and commit sha. If the commit
// has no associated pull requests than an empty slice is returned.
func GetPullRequests(ctx context.Context, client *githubv4.Client, githubOrg, repository, commitSha string) ([]*PullRequest, error) {
	var query struct {
		Repository struct {
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
	return pullRequests, nil
}
