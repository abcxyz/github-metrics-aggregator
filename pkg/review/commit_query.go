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
	"fmt"
	"strings"
	"text/template"
)

// commitSQL is the BigQuery query that selects the commits that need
// to be processed. The criteria for a commit that needs to be processed are:
// 1. The commit was pushed to the repository's default branch.
// 2. We do not have a record for the commit in the commit_review_status table.
const commitSQL = `
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
    {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.PushEventsTableID}}{{.BT}} push_events,
    UNNEST(push_events.commits) commit_json
  WHERE
    push_events.ref = CONCAT('refs/heads/', push_events.repository_default_branch)
    AND push_events.compare_url LIKE '{{.GitHubURLPrefix}}%' )
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
  {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.CommitReviewStatusTableID}}{{.BT}} commit_review_status
ON
  commit_review_status.commit_sha = commits.commit_sha
WHERE
  commit_review_status.commit_sha IS NULL
`

type queryParameters struct {
	ProjectID                 string
	DatasetID                 string
	PushEventsTableID         string
	CommitReviewStatusTableID string
	GitHubURLPrefix           string
	BT                        string
}

// makeCommitQuery returns a BigQuery query that selects the commits that need to be
// processed.
func makeCommitQuery(cfg *Config) (string, error) {
	tmpl, err := template.New("commit-query").Parse(commitSQL)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &queryParameters{
		ProjectID:                 cfg.ProjectID,
		DatasetID:                 cfg.DatasetID,
		PushEventsTableID:         cfg.PushEventsTableID,
		CommitReviewStatusTableID: cfg.CommitReviewStatusTableID,
		GitHubURLPrefix:           ghURLPrefix(cfg.GitHub.GitHubEnterpriseServerURL),
		BT:                        "`",
	}); err != nil {
		return "", fmt.Errorf("failed to apply query template parameters: %w", err)
	}
	return sb.String(), nil
}

func ghURLPrefix(url string) string {
	ghURLPrefix := url
	if ghURLPrefix == "" {
		ghURLPrefix = "https://github.com"
	}
	return ghURLPrefix + "/"
}
