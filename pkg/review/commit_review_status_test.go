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

package review

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/testutil"
)

func TestGetPullRequests(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		token             string
		githubOrg         string
		repository        string
		commitSha         string
		responseBodies    []string
		wantRequestBodies []string
		want              []*PullRequest
		wantErr           string
	}{
		{
			name:       "one_pull_request_with_one_page",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
			 "reviewCursor": null
           }
         }`,
			},
			want: []*PullRequest{
				{
					BaseRefName: "main",
					DatabaseID:  1,
					Number:      23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "XQ",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 1
                 }
               }
             }
           }
         }`,
			},
		},
		{
			name:       "two_pull_requests_with_one_page",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
			},
			want: []*PullRequest{
				{
					BaseRefName: "main",
					DatabaseID:  1,
					Number:      23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
				{
					BaseRefName: "main",
					DatabaseID:  2,
					Number:      48,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
					URL: "https://github.com/my-org/my-repo/pull/48",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     },
                     {
                       "baseRefName": "main",
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": []
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "XQ",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": "ER"
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
			},
		},
		{
			name:       "two_pull_requests_with_two_pages",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "XQ",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
			},
			want: []*PullRequest{
				{
					BaseRefName: "main",
					DatabaseID:  1,
					Number:      23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
				{
					BaseRefName: "main",
					DatabaseID:  2,
					Number:      48,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
					URL: "https://github.com/my-org/my-repo/pull/48",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "XQ",
                     "hasNextPage": true,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": []
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "FG",
                     "hasNextPage": false,
                     "hasPreviousPage": true,
                     "startCursor": "XQ"
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
			},
		},
		{
			name:       "no_associated_pull_requests_for_a_commit",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
			},
			want: []*PullRequest{},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [],
                   "pageInfo": {
                     "endCursor": "XQ",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 0
                 }
               }
             }
           }
         }`,
			},
		},
		{
			name:       "pull_requests_not_merged_to_default_branch_filtered_out",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo {
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo {
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
			},
			want: []*PullRequest{
				{
					BaseRefName: "main",
					DatabaseID:  1,
					Number:      23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     },
                     {
                       "baseRefName": "feature-branch",
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": []
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "XQ",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": "ER"
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
			},
		},
		{
			name:       "one_pull_request_with_two_pages_of_reviews",
			token:      "fake_token",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": null
           }
         }`,
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pullRequestCursor:String! $repository:String! $reviewCursor:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 defaultBranchRef {
                   name
                 },
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pullRequestCursor) {
                       nodes{
                         baseRefName,
                         databaseId,
                         number,
                         reviews(first: 100, after: $reviewCursor) {
                           nodes {
                             state
                           },
                           pageInfo{
                             hasNextPage,
                             hasPreviousPage,
                             endCursor,
                             startCursor
                           }
                         },
                         url
                       },
                       pageInfo{
                         hasNextPage,
                         hasPreviousPage,
                         endCursor,
                         startCursor
                       },
                       totalCount
                     }
                   }
                 }
               }
             }
           ",
           "variables": {
             "commitSha": "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
             "githubOrg": "test-org",
             "pullRequestCursor": "",
             "repository":"test-repo",
						 "reviewCursor": "XQ"
           }
         }`,
			},
			want: []*PullRequest{
				{
					BaseRefName: "main",
					DatabaseID:  1,
					Number:      23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "CHANGES_REQUESTED",
							},
							{
								State: "APPROVED",
							},
						},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "CHANGES_REQUESTED"
                           }
                         ],
                         "pageInfo": {
                           "hasNextPage": true,
                           "hasPreviousPage": false,
                           "endCursor": "XQ",
                           "startCursor": ""
                         }
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
				`{
           "data": {
             "repository": {
               "defaultBranchRef": {
                 "name": "main"
               },
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "baseRefName": "main",
                       "databaseId": 1,
                       "number": 23,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
                       },
                       "url": "https://github.com/my-org/my-repo/pull/23"
                     }
                   ],
                   "pageInfo": {
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "endCursor": "",
                     "startCursor": ""
                   },
                   "totalCount": 2
                 }
               }
             }
           }
         }`,
			},
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotRequestBodies := make([]string, 0)
			requestNumber := 0
			fakeGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedAuthHeader := fmt.Sprintf("Bearer %s", tc.token)
				if r.Header.Get("Authorization") != expectedAuthHeader {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}
				bytes, _ := io.ReadAll(r.Body)
				requestBody := string(bytes)
				gotRequestBodies = append(gotRequestBodies, requestBody)
				fmt.Fprintf(w, tc.responseBodies[requestNumber])
				requestNumber++
			}))
			src := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: tc.token},
			)
			ctx := context.Background()
			httpClient := oauth2.NewClient(ctx, src)
			client := githubv4.NewEnterpriseClient(fakeGitHub.URL, httpClient)
			got, err := GetPullRequestsTargetingDefaultBranch(ctx, client, tc.githubOrg, tc.repository, tc.commitSha)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("GetPullRequestsTargetingDefaultBranch got unexpected result (-got,+want):\n%s", diff)
			}
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}
			if diff := cmp.Diff(normalize(gotRequestBodies), normalize(tc.wantRequestBodies)); diff != "" {
				t.Errorf("Incorrect Request Bodies (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestGetCommitQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                    string
		pushEventsTable         *bigqueryio.QualifiedTableName
		commitReviewStatusTable *bigqueryio.QualifiedTableName
		want                    string
	}{
		{
			name: "query_template_populated_correctly",
			pushEventsTable: &bigqueryio.QualifiedTableName{
				Project: "my_project",
				Dataset: "my_dataset",
				Table:   "push_events",
			},
			commitReviewStatusTable: &bigqueryio.QualifiedTableName{
				Project: "my_project",
				Dataset: "my_dataset",
				Table:   "commit_review_status",
			},
			want: `
WITH
  commits AS (
  SELECT
    push_events.pusher author,
    push_events.organization,
    push_events.repository,
    push_events.repository_default_branch branch,
    JSON_VALUE(commit_json, '$.id') commit_sha,
    JSON_VALUE(commit_json, '$.timestamp') commit_timestamp,
  FROM
    ` + "`my_project.my_dataset.push_events`" + ` push_events,
    UNNEST(push_events.commits) commit_json
  WHERE
    push_events.ref = CONCAT('refs/heads/', push_events.repository_default_branch) )
SELECT
  commits.author,
  commits.organization,
  commits.repository,
  commits.branch,
  commits.commit_sha,
  commits.commit_timestamp
FROM
  commits
LEFT JOIN
  ` + "`my_project.my_dataset.commit_review_status`" + ` commit_review_status
ON
  commit_review_status.commit_sha = commits.commit_sha
WHERE
  commit_review_status.commit_sha IS NULL
`,
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := commitQuery(tc.pushEventsTable, tc.commitReviewStatusTable)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("GetCommitQuery got unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestGetBreakGlassIssueQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		issuesTable *bigqueryio.QualifiedTableName
		user        string
		timestamp   string
		want        string
	}{
		{
			name: "query_template_populated_correctly",
			issuesTable: &bigqueryio.QualifiedTableName{
				Project: "my_project",
				Dataset: "my_dataset",
				Table:   "issues",
			},
			user:      "bbechtel",
			timestamp: "2023-08-15T23:21:34Z",
			want: `
SELECT
  issues.html_url html_url
FROM
  ` + "`my_project.my_dataset.issues`" + ` issues
WHERE
  issues.repository = 'breakglass'
  AND author = 'bbechtel'
  AND issues.created_at <= TIMESTAMP('2023-08-15T23:21:34Z')
  AND issues.closed_at >= TIMESTAMP('2023-08-15T23:21:34Z')
`,
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := breakGlassIssueQuery(tc.issuesTable, tc.user, tc.timestamp)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("GetBreakGlassIssueQuery unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestGetPullRequest(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		pullRequests []*PullRequest
		want         *PullRequest
	}{
		{
			name: "returns_first_approving_pull_request",
			pullRequests: []*PullRequest{
				{
					DatabaseID: 1,
					Number:     23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
				},
				{
					DatabaseID: 2,
					Number:     24,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
				},
				{
					DatabaseID: 5,
					Number:     345,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
				},
			},
			want: &PullRequest{
				DatabaseID: 1,
				Number:     23,
				Reviews: struct {
					Nodes    []*Review
					PageInfo PageInfo
				}{
					Nodes: []*Review{
						{
							State: "APPROVED",
						},
					},
				},
			},
		},
		{
			name: "returns_nil_when_no_approving_pull_requests",
			pullRequests: []*PullRequest{
				{
					DatabaseID: 1,
					Number:     23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
				},
				{
					DatabaseID: 2,
					Number:     24,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
				},
				{
					DatabaseID: 5,
					Number:     345,
					Reviews: struct {
						Nodes    []*Review
						PageInfo PageInfo
					}{
						Nodes: []*Review{},
					},
				},
			},
			want: nil,
		},
		{
			name:         "returns_nil_when_no_pull_requests",
			pullRequests: []*PullRequest{},
			want:         nil,
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := getApprovingPullRequest(tc.pullRequests)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("getCommitHTMLURL unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestGetCommitHtmlUrl(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		commit Commit
		want   string
	}{
		{
			name: "url_template_populated_correctly",
			commit: Commit{
				Organization: "test-org",
				Repository:   "test-repo",
				SHA:          "123456",
			},
			want: "https://github.com/test-org/test-repo/commit/123456",
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := getCommitHTMLURL(tc.commit)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("getCommitHTMLURL unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestCommitApprovalDoFn_ProcessElement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                string
		token               string
		graphQlResponseCode int
		graphQLResponse     string
		config              *CommitApprovalPipelineConfig
		commit              Commit
		want                CommitReviewStatus
	}{
		{
			name:  "converts_commit_to_commit_review_status_correctly",
			token: "fake-token",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "FG",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 1
                 }
               }
             }
           }
         }`,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					SHA:          "12345678",
					Timestamp:    "2023-10-06T14:22:33Z",
				},
				HTMLURL:            "https://github.com/test-org/test-repository/commit/12345678",
				PullRequestID:      2,
				PullRequestNumber:  48,
				PullRequestHTMLURL: "https://github.com/my-org/my-repo/pull/48",
				ApprovalStatus:     GithubPRApproved,
				BreakGlassURLs:     []string{},
			},
		},
		{
			name:  "commit_considered_approved_as_long_as_one_pr_approves",
			token: "fake-token",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "CHANGES_REQUESTED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     },
                     {
                       "databaseId": 3,
                       "number": 52,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "APPROVED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/52"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "FG",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 1
                 }
               }
             }
           }
         }`,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					SHA:          "12345678",
					Timestamp:    "2023-10-06T14:22:33Z",
				},
				HTMLURL:            "https://github.com/test-org/test-repository/commit/12345678",
				PullRequestID:      3,
				PullRequestNumber:  52,
				PullRequestHTMLURL: "https://github.com/my-org/my-repo/pull/52",
				ApprovalStatus:     GithubPRApproved,
				BreakGlassURLs:     []string{},
			},
		},
		{
			name:  "uses_first_pr_if_no_prs_approve",
			token: "fake-token",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 2,
                       "number": 48,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "CHANGES_REQUESTED"
                           }
                         ]
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     },
                     {
                       "databaseId": 3,
                       "number": 52,
                       "reviews": {
                         "nodes": []
                       },
                       "url": "https://github.com/my-org/my-repo/pull/52"
                     }
                   ],
                   "pageInfo": {
                     "endCursor": "FG",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 1
                 }
               }
             }
           }
         }`,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					SHA:          "12345678",
					Timestamp:    "2023-10-06T14:22:33Z",
				},
				HTMLURL:            "https://github.com/test-org/test-repository/commit/12345678",
				PullRequestID:      2,
				PullRequestNumber:  48,
				PullRequestHTMLURL: "https://github.com/my-org/my-repo/pull/48",
				ApprovalStatus:     "CHANGES_REQUESTED",
				BreakGlassURLs:     []string{},
			},
		},
		{
			name: "default_approval_status_assigned_when_no_associated_prs",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			token:               "fake-token",
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [],
                   "pageInfo": {
                     "endCursor": "FG",
                     "hasNextPage": false,
                     "hasPreviousPage": false,
                     "startCursor": ""
                   },
                   "totalCount": 0
                 }
               }
             }
           }
         }`,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					SHA:          "12345678",
					Timestamp:    "2023-10-06T14:22:33Z",
				},
				HTMLURL:        "https://github.com/test-org/test-repository/commit/12345678",
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{},
			},
		},
		{
			name: "nothing_emitted_when_error_getting_prs",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{},
		},
		{
			name: "failed_to_get_repository_emitted_with_note",
			config: &CommitApprovalPipelineConfig{
				PushEventsTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "push_events",
				},
				CommitReviewStatusTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "commit_review_status",
				},
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			token:               "fake-token",
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {},
           "errors": [
             {
               "message": "Could not resolve to a Repository with the name 'test-repository'"
             }
            ]
         }`,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				SHA:          "12345678",
				Timestamp:    "2023-10-06T14:22:33Z",
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					SHA:          "12345678",
					Timestamp:    "2023-10-06T14:22:33Z",
				},
				HTMLURL:        "https://github.com/test-org/test-repository/commit/12345678",
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{},
				Note:           "Could not resolve to a Repository with the name 'test-repository'",
			},
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fakeGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedAuthHeader := fmt.Sprintf("Bearer %s", tc.token)
				if r.Header.Get("Authorization") != expectedAuthHeader {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}
				w.WriteHeader(tc.graphQlResponseCode)
				fmt.Fprintf(w, tc.graphQLResponse)
			}))
			src := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: tc.token},
			)
			ctx := context.Background()
			httpClient := oauth2.NewClient(ctx, src)
			client := githubv4.NewEnterpriseClient(fakeGitHub.URL, httpClient)
			commitApprovalDoFn := CommitApprovalDoFn{
				Config:       tc.config,
				githubClient: client,
			}
			var got CommitReviewStatus
			commitApprovalDoFn.ProcessElement(ctx, tc.commit, func(status CommitReviewStatus) {
				got = status
			})
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("commitApprovalDoFn.ProcessElement unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

func TestBreakGlassIssueDoFn_ProcessElement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name               string
		config             *CommitApprovalPipelineConfig
		commitReviewStatus CommitReviewStatus
		testFetcher        func(context.Context, string, string) ([]*breakGlassIssue, error)
		issueTable         *bigqueryio.QualifiedTableName
		author             string
		timestamp          string
		want               CommitReviewStatus
	}{
		{
			name: "break_glass_url_loads_if_bigquery_returns_successfully",
			config: &CommitApprovalPipelineConfig{
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					SHA:          "12345",
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error) {
				return []*breakGlassIssue{
					{HTMLURL: "https://github.com/test-org/breakglass/issues/5"},
				}, nil
			},
			issueTable: &bigqueryio.QualifiedTableName{
				Project: "test-project",
				Dataset: "test-dataset",
				Table:   "test-table",
			},
			author:    "bbechtel",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					SHA:          "12345",
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{"https://github.com/test-org/breakglass/issues/5"},
			},
		},
		{
			name: "multiple_break_glass_issues_are_supported",
			config: &CommitApprovalPipelineConfig{
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					SHA:          "12345",
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error) {
				return []*breakGlassIssue{
					{HTMLURL: "https://github.com/test-org/breakglass/issues/5"},
					{HTMLURL: "https://github.com/test-org/breakglass/issues/6"},
				}, nil
			},
			issueTable: &bigqueryio.QualifiedTableName{
				Project: "test-project",
				Dataset: "test-dataset",
				Table:   "test-table",
			},
			author:    "bbechtel",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					SHA:          "12345",
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{
					"https://github.com/test-org/breakglass/issues/5",
					"https://github.com/test-org/breakglass/issues/6",
				},
			},
		},
		{
			name: "nothing_emitted_when_bigquery_returns_error",
			config: &CommitApprovalPipelineConfig{
				IssuesTable: &bigqueryio.QualifiedTableName{
					Project: "test-project",
					Dataset: "test-dataset",
					Table:   "issues",
				},
			},
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					SHA:          "12345",
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error) {
				return nil, errors.New("bigquery unavailable")
			},
			issueTable: &bigqueryio.QualifiedTableName{
				Project: "test-project",
				Dataset: "test-dataset",
				Table:   "test-table",
			},
			author:    "bbechtel",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			want:      CommitReviewStatus{},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			client := &TestBreakGlassIssueFetcher{
				fetcher: tc.testFetcher,
			}
			breakGlassIssueDoFn := BreakGlassIssueDoFn{
				Config:                 tc.config,
				breakGlassIssueFetcher: client,
			}
			var got CommitReviewStatus
			breakGlassIssueDoFn.ProcessElement(ctx, tc.commitReviewStatus, func(status CommitReviewStatus) {
				got = status
			})
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("breakGlassIssueDoFn.ProcessElement got unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}

type TestBreakGlassIssueFetcher struct {
	fetcher func(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error)
}

func (tbgif *TestBreakGlassIssueFetcher) getBreakGlassIssues(ctx context.Context, author, timestamp string) ([]*breakGlassIssue, error) {
	return tbgif.fetcher(ctx, author, timestamp)
}

func normalize(strings []string) []string {
	normalized := make([]string, 0, len(strings))
	for _, s := range strings {
		normalized = append(normalized, removeWhiteSpace(s))
	}
	return normalized
}

func removeWhiteSpace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
