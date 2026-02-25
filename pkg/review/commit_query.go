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
    JSON_VALUE(payload, '$.pusher.name') author,
    JSON_VALUE(payload, '$.organization.login') organization,
    JSON_VALUE(payload, '$.repository.name') repository,
    JSON_VALUE(payload, '$.repository.default_branch') branch,
    JSON_VALUE(payload, '$.repository.visibility') visibility,
    JSON_VALUE(commit_json, '$.id') commit_sha,
    TIMESTAMP(JSON_VALUE(commit_json, '$.timestamp')) commit_timestamp,
  FROM
    {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.EventsTableID}}{{.BT}} events,
    UNNEST(JSON_EXTRACT_ARRAY(payload, '$.commits')) commit_json
  WHERE
    event = 'PushEvent'
    AND JSON_VALUE(payload, '$.ref') = CONCAT('refs/heads/', JSON_VALUE(payload, '$.repository.default_branch'))
    AND JSON_VALUE(payload, '$.compare_url') LIKE '{{.GitHubURLPrefix}}%' )
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
	EventsTableID             string
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
		EventsTableID:             cfg.EventsTableID,
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
