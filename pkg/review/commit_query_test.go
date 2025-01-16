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
    push_events.pusher author,
    push_events.organization,
    push_events.repository,
    push_events.repository_default_branch branch,
    push_events.repository_visibility visibility,
    JSON_VALUE(commit_json, '$.id') commit_sha,
    TIMESTAMP(JSON_VALUE(commit_json, '$.timestamp')) commit_timestamp,
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
