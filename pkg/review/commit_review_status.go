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
)

// PullRequest represents a pull request in GitHub and contains the
// GitHub assigned ID, the pull request number in the repository,
// and the review decision for the pull request.
type PullRequest struct {
	DatabaseID     githubv4.Int
	Number         githubv4.Int
	ReviewDecision githubv4.String
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
