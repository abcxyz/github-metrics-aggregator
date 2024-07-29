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

	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/testutil"
)

var defaultConfig = &Config{
	ProjectID:                 "my_project",
	DatasetID:                 "my_dataset",
	PushEventsTableID:         "push_events",
	CommitReviewStatusTableID: "commit_review_status",
	IssuesTableID:             "issues",
}

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
                         fullDatabaseId,
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
					BaseRefName:    "main",
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
						PageInfo: &PageInfo{},
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
                       "fullDatabaseId": "1",
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
                         fullDatabaseId,
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
					BaseRefName:    "main",
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
						PageInfo: &PageInfo{},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
				{
					BaseRefName:    "main",
					FullDatabaseID: "2",
					Number:         48,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes:    []*Review{},
						PageInfo: &PageInfo{},
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
                       "fullDatabaseId": "1",
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
                     },
                     {
                       "baseRefName": "main",
                       "fullDatabaseId": "2",
                       "number": 48,
                       "reviews": {
                         "nodes": [],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
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
                         fullDatabaseId,
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
                         fullDatabaseId,
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
					BaseRefName:    "main",
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
						PageInfo: &PageInfo{},
					},
					URL: "https://github.com/my-org/my-repo/pull/23",
				},
				{
					BaseRefName:    "main",
					FullDatabaseID: "2",
					Number:         48,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes:    []*Review{},
						PageInfo: &PageInfo{},
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
                       "fullDatabaseId": "1",
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
                       "fullDatabaseId": "2",
                       "number": 48,
                       "reviews": {
                         "nodes": [],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
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
                         fullDatabaseId,
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
                         fullDatabaseId,
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
					BaseRefName:    "main",
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
						PageInfo: &PageInfo{},
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
                       "fullDatabaseId": "1",
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
                     },
                     {
                       "baseRefName": "feature-branch",
                       "fullDatabaseId": "2",
                       "number": 48,
                       "reviews": {
                         "nodes": [],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
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
                         fullDatabaseId,
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
                         fullDatabaseId,
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
					BaseRefName:    "main",
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "CHANGES_REQUESTED",
							},
							{
								State: "APPROVED",
							},
						},
						PageInfo: &PageInfo{},
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
                       "fullDatabaseId": "1",
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
                       "fullDatabaseId": "1",
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
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
				},
				{
					FullDatabaseID: "2",
					Number:         24,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{
							{
								State: "APPROVED",
							},
						},
					},
				},
				{
					FullDatabaseID: "5",
					Number:         345,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{},
					},
				},
			},
			want: &PullRequest{
				FullDatabaseID: "1",
				Number:         23,
				Reviews: struct {
					Nodes    []*Review
					PageInfo *PageInfo
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
					FullDatabaseID: "1",
					Number:         23,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{},
					},
				},
				{
					FullDatabaseID: "2",
					Number:         24,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
					}{
						Nodes: []*Review{},
					},
				},
				{
					FullDatabaseID: "5",
					Number:         345,
					Reviews: struct {
						Nodes    []*Review
						PageInfo *PageInfo
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

func TestProcessCommit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                string
		token               string
		graphQlResponseCode int
		graphQLResponse     string
		cfg                 *Config
		commit              Commit
		want                CommitReviewStatus
	}{
		{
			name:                "converts_commit_to_commit_review_status_correctly",
			token:               "fake-token",
			cfg:                 defaultConfig,
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "fullDatabaseId": "2",
                       "number": 48,
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
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					Visibility:   "public",
					SHA:          "12345678",
					Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
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
			name:                "commit_considered_approved_as_long_as_one_pr_approves",
			token:               "fake-token",
			cfg:                 defaultConfig,
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "fullDatabaseId": "2",
                       "number": 48,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "CHANGES_REQUESTED"
                           }
                         ],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     },
                     {
                       "fullDatabaseId": "3",
                       "number": 52,
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
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					Visibility:   "public",
					SHA:          "12345678",
					Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
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
			name:                "uses_first_pr_if_no_prs_approve",
			token:               "fake-token",
			cfg:                 defaultConfig,
			graphQlResponseCode: 200,
			graphQLResponse: `{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "fullDatabaseId": "2",
                       "number": 48,
                       "reviews": {
                         "nodes": [
                           {
                             "state": "CHANGES_REQUESTED"
                           }
                         ],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
                       },
                       "url": "https://github.com/my-org/my-repo/pull/48"
                     },
                     {
                       "fullDatabaseId": "3",
                       "number": 52,
                       "reviews": {
                         "nodes": [],
                         "pageInfo": {
                           "hasNextPage": false,
                           "hasPreviousPage": false,
                           "endCursor": "",
                           "startCursor": ""
                         }
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
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					Visibility:   "public",
					SHA:          "12345678",
					Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
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
			name:                "default_approval_status_assigned_when_no_associated_prs",
			cfg:                 defaultConfig,
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
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					Visibility:   "public",
					SHA:          "12345678",
					Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
				},
				HTMLURL:        "https://github.com/test-org/test-repository/commit/12345678",
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{},
			},
		},
		{
			name: "nothing_emitted_when_error_getting_prs",
			cfg:  defaultConfig,
			commit: Commit{
				Author:       "test-author",
				Organization: "test-org",
				Repository:   "test-repository",
				Branch:       "main",
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{},
		},
		{
			name:                "failed_to_get_repository_emitted_with_note",
			cfg:                 defaultConfig,
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
				Visibility:   "public",
				SHA:          "12345678",
				Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
			},
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repository",
					Branch:       "main",
					Visibility:   "public",
					SHA:          "12345678",
					Timestamp:    time.Date(2023, 10, 6, 14, 22, 33, 0, time.UTC),
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
			got := processCommit(ctx, tc.commit, client)
			if got != nil {
				if diff := cmp.Diff(*got, tc.want); diff != "" {
					t.Errorf("processCommit: unexpected result (-got,+want):\n%s", diff)
				}
			}
		})
	}
}

func TestProcessReviewStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name               string
		cfg                *Config
		commitReviewStatus CommitReviewStatus
		testFetcher        func(context.Context, string, *time.Time) ([]*breakGlassIssue, error)
		author             string
		timestamp          string
		want               CommitReviewStatus
	}{
		{
			name: "break_glass_url_loads_if_bigquery_returns_successfully",
			cfg:  defaultConfig,
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					Visibility:   "public",
					SHA:          "12345",
					Timestamp:    time.Date(2024, 7, 12, 10, 20, 17, 70, time.UTC),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author string, timestamp *time.Time) ([]*breakGlassIssue, error) {
				return []*breakGlassIssue{
					{HTMLURL: "https://github.com/test-org/breakglass/issues/5"},
				}, nil
			},
			author:    "bbechtel",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					Visibility:   "public",
					SHA:          "12345",
					Timestamp:    time.Date(2024, 7, 12, 10, 20, 17, 70, time.UTC),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
				BreakGlassURLs: []string{"https://github.com/test-org/breakglass/issues/5"},
			},
		},
		{
			name: "multiple_break_glass_issues_are_supported",
			cfg:  defaultConfig,
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					Visibility:   "public",
					SHA:          "12345",
					Timestamp:    time.Date(2024, 7, 12, 10, 20, 17, 70, time.UTC),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author string, timestamp *time.Time) ([]*breakGlassIssue, error) {
				return []*breakGlassIssue{
					{HTMLURL: "https://github.com/test-org/breakglass/issues/5"},
					{HTMLURL: "https://github.com/test-org/breakglass/issues/6"},
				}, nil
			},
			author:    "bbechtel",
			timestamp: time.Now().UTC().Format(time.RFC3339),
			want: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					Visibility:   "public",
					SHA:          "12345",
					Timestamp:    time.Date(2024, 7, 12, 10, 20, 17, 70, time.UTC),
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
			cfg:  defaultConfig,
			commitReviewStatus: CommitReviewStatus{
				Commit: Commit{
					Author:       "test-author",
					Organization: "test-org",
					Repository:   "test-repo",
					Branch:       "test-branch",
					Visibility:   "public",
					SHA:          "12345",
					Timestamp:    time.Date(2024, 7, 12, 10, 20, 17, 70, time.UTC),
				},
				HTMLURL:        "",
				PullRequestID:  0,
				ApprovalStatus: DefaultApprovalStatus,
			},
			testFetcher: func(ctx context.Context, author string, timestamp *time.Time) ([]*breakGlassIssue, error) {
				return nil, errors.New("bigquery unavailable")
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
			fetcher := &TestBreakGlassIssueFetcher{
				fetcher: tc.testFetcher,
			}
			got := processReviewStatus(ctx, fetcher, tc.cfg, tc.commitReviewStatus)
			if got != nil {
				if diff := cmp.Diff(*got, tc.want); diff != "" {
					t.Errorf("proecessReviewStatus: got unexpected result (-got,+want):\n%s", diff)
				}
			}
		})
	}
}

type TestBreakGlassIssueFetcher struct {
	fetcher func(ctx context.Context, author string, timestamp *time.Time) ([]*breakGlassIssue, error)
}

func (tbgif *TestBreakGlassIssueFetcher) fetch(ctx context.Context, cfg *Config, author string, timestamp *time.Time) ([]*breakGlassIssue, error) {
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
