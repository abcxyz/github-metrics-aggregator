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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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
			name:       "One Pull Request With One Page",
			token:      "ghp_JQqAJky0GlNB7xqVCbUrivIgFZ09V8gGCi5C",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pageCursor:String! $repository:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pageCursor) {
                       nodes{
                         databaseId,
                         number,
                         reviewDecision
                       },
                       pageInfo{
                         endCursor,
                         hasNextPage,
                         hasPreviousPage,
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
             "pageCursor": "",
             "repository":"test-repo"
           }
         }`,
			},
			want: []*PullRequest{
				{
					DatabaseID:     1,
					Number:         23,
					ReviewDecision: "APPROVED",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 1,
                       "number": 23,
                       "reviewDecision": "APPROVED"
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
			name:       "Two Pull Requests With One Page",
			token:      "ghp_JQqAJky0GlNB7xqVCbUrivIgFZ09V8gGCi5C",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pageCursor:String! $repository:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pageCursor) {
                       nodes{
                         databaseId,
                         number,
                         reviewDecision
                       },
                       pageInfo{
                         endCursor,
                         hasNextPage,
                         hasPreviousPage,
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
             "pageCursor": "",
             "repository":"test-repo"
           }
         }`,
			},
			want: []*PullRequest{
				{
					DatabaseID:     1,
					Number:         23,
					ReviewDecision: "APPROVED",
				},
				{
					DatabaseID:     2,
					Number:         48,
					ReviewDecision: "REVIEW_REQUESTED",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 1,
                       "number": 23,
                       "reviewDecision": "APPROVED"
                     },
                     {
                       "databaseId": 2,
                       "number": 48,
                       "reviewDecision": "REVIEW_REQUESTED"
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
			name:       "Two Pull Requests With Two Pages",
			token:      "ghp_JQqAJky0GlNB7xqVCbUrivIgFZ09V8gGCi5C",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pageCursor:String! $repository:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pageCursor) {
                       nodes{
                         databaseId,
                         number,
                         reviewDecision
                       },
                       pageInfo{
                         endCursor,
                         hasNextPage,
                         hasPreviousPage,
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
             "pageCursor": "",
             "repository":"test-repo"
           }
         }`,
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pageCursor:String! $repository:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pageCursor) {
                       nodes{
                         databaseId,
                         number,
                         reviewDecision
                       },
                       pageInfo{
                         endCursor,
                         hasNextPage,
                         hasPreviousPage,
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
             "pageCursor": "XQ",
             "repository":"test-repo"
           }
         }`,
			},
			want: []*PullRequest{
				{
					DatabaseID:     1,
					Number:         23,
					ReviewDecision: "APPROVED",
				},
				{
					DatabaseID:     2,
					Number:         48,
					ReviewDecision: "REVIEW_REQUESTED",
				},
			},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 1,
                       "number": 23,
                       "reviewDecision": "APPROVED"
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
               "object": {
                 "associatedPullRequests": {
                   "nodes": [
                     {
                       "databaseId": 2,
                       "number": 48,
                       "reviewDecision": "REVIEW_REQUESTED"
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
			name:       "No Associated Pull Requests for a commit",
			token:      "ghp_JQqAJky0GlNB7xqVCbUrivIgFZ09V8gGCi5C",
			githubOrg:  "test-org",
			repository: "test-repo",
			commitSha:  "kof6p96lr6qvdu81qw49fhmoxrod9qmc2qak51nh",
			wantRequestBodies: []string{
				`{
           "query": "
             query($commitSha:GitObjectID! $githubOrg:String! $pageCursor:String! $repository:String!) {
               repository(owner: $githubOrg, name: $repository) {
                 object(oid: $commitSha) {
                   ... on Commit{
                     associatedPullRequests(first: 100, after: $pageCursor) {
                       nodes{
                         databaseId,
                         number,
                         reviewDecision
                       },
                       pageInfo{
                         endCursor,
                         hasNextPage,
                         hasPreviousPage,
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
             "pageCursor": "",
             "repository":"test-repo"
           }
         }`,
			},
			want: []*PullRequest{},
			responseBodies: []string{
				`{
           "data": {
             "repository": {
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
	}
	for _, tc := range cases {
		requestNumber := 0
		fakeGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedAuthHeader := fmt.Sprintf("Bearer %s", tc.token)
			if r.Header.Get("Authorization") != expectedAuthHeader {
				w.WriteHeader(500)
				fmt.Fprintf(w, "missing or malformed authorization header")
				return
			}
			bytes, _ := io.ReadAll(r.Body)
			body := string(bytes)
			wantBody := tc.wantRequestBodies[requestNumber]
			if !equalExceptWhiteSpace(body, wantBody) {
				t.Errorf("Incorrect Request Body: got=%s, want=%s", body, wantBody)
			}
			fmt.Fprintf(w, tc.responseBodies[requestNumber])
			requestNumber++
		}))
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: tc.token},
		)
		httpClient := oauth2.NewClient(context.Background(), src)
		client := githubv4.NewEnterpriseClient(fakeGitHub.URL, httpClient)
		got, err := GetPullRequests(context.Background(), client, tc.githubOrg, tc.repository, tc.commitSha)
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("GetPullRequests got unexpected result (-got,+want):\n%s", diff)
		}
		if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
			t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
		}
	}
}

func equalExceptWhiteSpace(s1, s2 string) bool {
	return removeWhiteSpace(s1) == removeWhiteSpace(s2)
}

func removeWhiteSpace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
