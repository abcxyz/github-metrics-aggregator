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

	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
)

func TestGetCommitQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "query_template_populated_correctly",
			cfg:  defaultConfig,
			want: `
WITH
  commits AS (
  SELECT
    JSON_VALUE(payload, '$.pusher.name') author,
    JSON_VALUE(payload, '$.organization.login') organization,
    JSON_VALUE(payload, '$.repository.name') repository,
    JSON_VALUE(payload, '$.repository.default_branch') branch,
    JSON_VALUE(payload, '$.repository.visibility') visibility,
    JSON_VALUE(commit_json, '$.id') commit_sha,
    TIMESTAMP(JSON_VALUE(commit_json, '$.timestamp')) commit_timestamp,
  FROM
    ` + "`my_project.my_dataset.events`" + ` events,
    UNNEST(JSON_EXTRACT_ARRAY(payload, '$.commits')) commit_json
  WHERE
    event = 'PushEvent'
    AND JSON_VALUE(payload, '$.ref') = CONCAT('refs/heads/', JSON_VALUE(payload, '$.repository.default_branch'))
    AND JSON_VALUE(payload, '$.compare_url') LIKE 'https://github.com/%' )
SELECT
  commits.author,
  commits.organization,
  commits.repository,
  commits.branch,
  commits.visibility,
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
		{
			name: "query_template_with_enterprise_url",
			cfg: &Config{
				GitHub: githubclient.Config{
					GitHubEnterpriseServerURL: "https://my-ghes.com",
				},
				ProjectID:                 "my_project",
				DatasetID:                 "my_dataset",
				EventsTableID:             "events",
				CommitReviewStatusTableID: "commit_review_status",
			},
			want: `
WITH
  commits AS (
  SELECT
    JSON_VALUE(payload, '$.pusher.name') author,
    JSON_VALUE(payload, '$.organization.login') organization,
    JSON_VALUE(payload, '$.repository.name') repository,
    JSON_VALUE(payload, '$.repository.default_branch') branch,
    JSON_VALUE(payload, '$.repository.visibility') visibility,
    JSON_VALUE(commit_json, '$.id') commit_sha,
    TIMESTAMP(JSON_VALUE(commit_json, '$.timestamp')) commit_timestamp,
  FROM
    ` + "`my_project.my_dataset.events`" + ` events,
    UNNEST(JSON_EXTRACT_ARRAY(payload, '$.commits')) commit_json
  WHERE
    event = 'PushEvent'
    AND JSON_VALUE(payload, '$.ref') = CONCAT('refs/heads/', JSON_VALUE(payload, '$.repository.default_branch'))
    AND JSON_VALUE(payload, '$.compare_url') LIKE 'https://my-ghes.com/%' )
SELECT
  commits.author,
  commits.organization,
  commits.repository,
  commits.branch,
  commits.visibility,
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, _ := makeCommitQuery(tc.cfg)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("GetCommitQuery got unexpected result (-got,+want):\n%s", diff)
			}
		})
	}
}
