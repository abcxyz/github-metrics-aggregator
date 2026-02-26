// Copyright 2024 The Authors (see AUTHORS file)
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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
)

func TestGetBreakGlassIssueQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		cfg          *Config
		user         string
		organization string
		timestamp    time.Time
		want         string
	}{
		{
			name:         "query_template_populated_correctly",
			cfg:          defaultConfig,
			user:         "bbechtel",
			organization: "test-org",
			timestamp:    time.Date(2023, 8, 15, 23, 21, 34, 0, time.UTC),
			want: `
SELECT
  JSON_VALUE(payload, '$.issue.html_url') html_url
FROM
  ` + "`my_project.my_dataset.optimized_events`" + ` events
WHERE
  event = 'issues'
  AND JSON_VALUE(payload, '$.repository.name') = 'breakglass'
  AND JSON_VALUE(payload, '$.organization.login') = 'test-org'
  AND JSON_VALUE(payload, '$.issue.user.login') = 'bbechtel'
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.created_at')) <= TIMESTAMP('2023-08-15T23:21:34Z')
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.closed_at')) >= TIMESTAMP('2023-08-15T23:21:34Z')
  AND JSON_VALUE(payload, '$.issue.html_url') LIKE 'https://github.com/%'
`,
		},
		{
			name: "query_template_with_enterprise_url",
			cfg: &Config{
				GitHub: githubclient.Config{
					GitHubEnterpriseServerURL: "https://my-ghes.com",
				},
				ProjectID:                 "my_project",
				DatasetID:                 "my_dataset",
				EventsTableID:             "optimized_events",
				CommitReviewStatusTableID: "commit_review_status",
			},
			user:         "test-user",
			organization: "test-org2",
			timestamp:    time.Date(2023, 8, 15, 23, 21, 34, 0, time.UTC),
			want: `
SELECT
  JSON_VALUE(payload, '$.issue.html_url') html_url
FROM
  ` + "`my_project.my_dataset.optimized_events`" + ` events
WHERE
  event = 'issues'
  AND JSON_VALUE(payload, '$.repository.name') = 'breakglass'
  AND JSON_VALUE(payload, '$.organization.login') = 'test-org2'
  AND JSON_VALUE(payload, '$.issue.user.login') = 'test-user'
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.created_at')) <= TIMESTAMP('2023-08-15T23:21:34Z')
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.closed_at')) >= TIMESTAMP('2023-08-15T23:21:34Z')
  AND JSON_VALUE(payload, '$.issue.html_url') LIKE 'https://my-ghes.com/%'
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := makeBreakglassQuery(tc.cfg, tc.user, tc.organization, &tc.timestamp)
			if err != nil {
				t.Errorf("unexpected error making breakglass query: %v", err)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("GetBreakGlassIssueQuery unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}
