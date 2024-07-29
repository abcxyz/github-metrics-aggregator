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
	"strconv"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/logging"
)

const (
	// GithubPRApproved is the approving review status indicating the PR
	// is approved to be merged.
	GithubPRApproved = "APPROVED"

	// GithubPRReviewRequired is the default review status of a PR indicating
	// the PR requires a review.
	GithubPRReviewRequired = "REVIEW_REQUIRED"

	// GithubPRChangesRequested is a blocking review status indicating that
	// changes need to be made to the PR code.
	GithubPRChangesRequested = "CHANGES_REQUESTED"

	// DefaultApprovalStatus is the default approval status we assign to a commit.
	DefaultApprovalStatus = "UNKNOWN"
)

// Commit maps the columns from the driving BigQuery query
// to a usable structure.
type Commit struct {
	Author       string `bigquery:"author"`
	Organization string `bigquery:"organization"`
	Repository   string `bigquery:"repository"`
	Branch       string `bigquery:"branch"`
	Visibility   string `bigquery:"visibility"`
	SHA          string `bigquery:"commit_sha"`

	// Timestamp will be in ISO 8601 format (https://en.wikipedia.org/wiki/ISO_8601)
	// and should be parsable using time.RFC3339 format
	Timestamp time.Time `bigquery:"commit_timestamp"`
}

// CommitReviewStatus maps the columns of the 'commit_review_status` table in
// BigQuery.
type CommitReviewStatus struct {
	Commit
	HTMLURL            string   `bigquery:"commit_html_url"`
	PullRequestID      int64    `bigquery:"pull_request_id"`
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

// CommitGraphQlQuery is struct that maps to the GitHub GraphQLQuery
// that fetches all the PRs and associated PR reviews for a commit sha.
type CommitGraphQlQuery struct {
	Repository struct {
		DefaultBranchRef struct {
			Name githubv4.String
		}
		Object struct {
			Commit struct {
				AssociatedPullRequest struct {
					Nodes      []*PullRequest
					PageInfo   *PageInfo
					TotalCount githubv4.Int
				} `graphql:"associatedPullRequests(first: 100, after: $pullRequestCursor)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(oid: $commitSha)"`
	} `graphql:"repository(owner: $githubOrg, name: $repository)"`
}

// PullRequest represents a pull request in GitHub and contains the
// GitHub assigned ID, the pull request number in the repository,
// and the review decision for the pull request.
// For all potential fields see:
// https://docs.github.com/en/graphql/reference/objects#pullrequest
type PullRequest struct {
	// BasRefName is the target the PR is being merged into. For example,
	// If a PR is being opened to merge the code from feature branch 'my-feature'
	// into branch 'main', then BasRefName for this PR would be 'main'.
	BaseRefName    githubv4.String
	FullDatabaseID githubv4.String
	Number         githubv4.Int
	Reviews        struct {
		Nodes    []*Review
		PageInfo *PageInfo
	} `graphql:"reviews(first: 100, after: $reviewCursor)"`
	URL githubv4.String
}

// Review represents a pull request review in GitHub's GraphQL API.
// For all potential fields see:
// https://docs.github.com/en/graphql/reference/objects#pullrequestreview
type Review struct {
	State githubv4.String
}

// PageInfo represents a pagination info in GitHub's GraphQL API.
// For all potential fields see:
// https://docs.github.com/en/graphql/reference/objects#pageinfo
type PageInfo struct {
	HasNextPage     githubv4.Boolean
	HasPreviousPage githubv4.Boolean
	EndCursor       githubv4.String
	StartCursor     githubv4.String
}

// processCommit is a function that takes a Commit, determines
// if the commit was properly approved, and returns the resulting
// CommitReviewStatus.
// A commit is considered properly reviewed as long as there is an associated
// PR for the commit targeting the repository's main branch with reviewDecision
// of 'APPROVED'.
func processCommit(ctx context.Context, commit Commit, gitHubClient *githubv4.Client) *CommitReviewStatus {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "process commit", "commit", commit)

	commitReviewStatus := CommitReviewStatus{
		Commit:         commit,
		HTMLURL:        getCommitHTMLURL(commit),
		ApprovalStatus: DefaultApprovalStatus,
		BreakGlassURLs: make([]string, 0),
	}
	requests, err := GetPullRequestsTargetingDefaultBranch(ctx, gitHubClient, commit.Organization, commit.Repository, commit.SHA)
	if err != nil {
		// Special error cases
		if strings.HasPrefix(err.Error(), "failed to call graphql") {
			unwrapped := errors.Unwrap(err)
			if strings.HasPrefix(unwrapped.Error(), "Could not resolve to a Repository") {
				// this is a permanent error from GitHub telling us the repository
				// for the commit no longer exists. Note this in the commit review status
				// and send it on for further processing
				commitReviewStatus.Note = unwrapped.Error()
				return &commitReviewStatus
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
		logger.ErrorContext(ctx, "failed to get pull requests for commit", "error", err)
		return nil // this commit could not be processed
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
		id, err := strconv.ParseInt(string(pullRequest.FullDatabaseID), 10, 64)
		if err != nil {
			// should never fail to parse as fullDatabaseId is of type BigInt
			// see: https://docs.github.com/en/graphql/reference/scalars#bigint
			panic("impossible")
		}
		commitReviewStatus.PullRequestID = id
		commitReviewStatus.PullRequestNumber = int(pullRequest.Number)
		commitReviewStatus.PullRequestHTMLURL = string(pullRequest.URL)
		commitReviewStatus.ApprovalStatus = getApprovalStatus(pullRequest)
	}
	return &commitReviewStatus
}

func getApprovalStatus(request *PullRequest) string {
	// All PRs start with status of GithubPRReviewRequired
	approvalStatus := GithubPRReviewRequired
	for _, review := range request.Reviews.Nodes {
		// if GithubPRChangesRequested set approvalStatus to that as we
		// want to know if a review was conducted but blocked the merge
		if review.State == GithubPRChangesRequested {
			approvalStatus = string(review.State)
		}
		// if GithubPRApproved is found immediately return as we know
		// the PR was approved and do not need to check other reviews.
		if review.State == GithubPRApproved {
			return GithubPRApproved
		}
	}
	return approvalStatus
}

// processReviewStatus is a function that takes a CommitReviewStatus
// and populates its breakGlassIssue field (if necessary) and then returns
// it. The process only searches for break glass
// issues for commits that do not have the status GithubPRApproved.
func processReviewStatus(ctx context.Context, fetcher BreakGlassIssueFetcher, cfg *Config, commitReviewStatus CommitReviewStatus) *CommitReviewStatus {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "processing commitReviewStatus", "commit_review_status", commitReviewStatus)
	if commitReviewStatus.ApprovalStatus != GithubPRApproved {
		// if the commit does not have proper approval, we check if there was a
		// break glass issue opened by the author during the timeframe they
		// submitted the commit.
		breakGlassIssues, err := fetcher.fetch(ctx, cfg, commitReviewStatus.Author, &commitReviewStatus.Timestamp)
		if err != nil {
			// We should only get transient style errors from BigQuery
			// (e.g. network is down etc.). So, we can just log the error and then
			// drop this CommitReviewStatus from the pipeline. It will then get
			// retried on the next run of the pipeline.
			logger.ErrorContext(ctx, "failure when trying to get break glass issue", "error", err)
			return nil
		}

		for _, v := range breakGlassIssues {
			commitReviewStatus.BreakGlassURLs = append(commitReviewStatus.BreakGlassURLs, v.HTMLURL)
		}
	}
	return &commitReviewStatus
}

// getApprovingPullRequest retrieves the first *PullRequest that has a
// review decision status with the value of GithubPRApproved. if no such
// *PullRequest is present then nil is returned.
func getApprovingPullRequest(pullRequests []*PullRequest) *PullRequest {
	for _, pullRequest := range pullRequests {
		for _, review := range pullRequest.Reviews.Nodes {
			if review.State == GithubPRApproved {
				return pullRequest
			}
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

// GetPullRequestsTargetingDefaultBranch retrieves all associated pull requests
// for a commit that target the repository's default branch from GitHub based on
// the given GitHub organization, repository, and commit sha. If the commit
// has no such associated pull requests then an empty slice is returned.
func GetPullRequestsTargetingDefaultBranch(ctx context.Context, client *githubv4.Client, githubOrg, repository, commitSha string) ([]*PullRequest, error) {
	var query CommitGraphQlQuery
	pullRequests := make([]*PullRequest, 0, query.Repository.Object.Commit.AssociatedPullRequest.TotalCount)
	pullRequestCursor := githubv4.String("")
	for {
		if err := client.Query(ctx, &query, map[string]interface{}{
			"githubOrg":         githubv4.String(githubOrg),
			"repository":        githubv4.String(repository),
			"commitSha":         githubv4.GitObjectID(commitSha),
			"pullRequestCursor": pullRequestCursor,
			// The initial reviewCursor must be nil and not the empty string "",
			// unlike the pullRequestCursor.
			"reviewCursor": (*githubv4.String)(nil),
		}); err != nil {
			return nil, fmt.Errorf("failed to call graphql: %w", err)
		}

		for i := 0; i < len(query.Repository.Object.Commit.AssociatedPullRequest.Nodes); i++ {
			pr := query.Repository.Object.Commit.AssociatedPullRequest.Nodes[i]
			if pr.BaseRefName == query.Repository.DefaultBranchRef.Name {
				// We need to account for when reviewNodes span multiple pages.
				for pr.Reviews.PageInfo.HasNextPage {
					// Make a new query object so that our existing query's
					// state is not obliterated.
					reviewQuery := CommitGraphQlQuery{}
					if err := client.Query(ctx, &reviewQuery, map[string]any{
						"githubOrg":         githubv4.String(githubOrg),
						"repository":        githubv4.String(repository),
						"commitSha":         githubv4.GitObjectID(commitSha),
						"pullRequestCursor": pullRequestCursor,
						"reviewCursor":      pr.Reviews.PageInfo.EndCursor,
					}); err != nil {
						return nil, fmt.Errorf("failed to call graphql: %w", err)
					}
					reviews := reviewQuery.Repository.Object.Commit.AssociatedPullRequest.Nodes[i].Reviews
					pr.Reviews.Nodes = append(pr.Reviews.Nodes, reviews.Nodes...)
					pr.Reviews.PageInfo = reviews.PageInfo
				}
				pullRequests = append(pullRequests, pr)
			}
		}
		pageInfo := query.Repository.Object.Commit.AssociatedPullRequest.PageInfo
		if pageInfo == nil || !pageInfo.HasNextPage {
			break
		}
		pullRequestCursor = query.Repository.Object.Commit.AssociatedPullRequest.PageInfo.EndCursor
	}
	return pullRequests, nil
}
