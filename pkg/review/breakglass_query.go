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
	"time"
)

// breakGlassIssueSQL is the BigQuery query that searches for a
// break glass issues created by given user and within a specified time frame.
const breakGlassIssueSQL = `
SELECT
  JSON_VALUE(payload, '$.issue.html_url') html_url
FROM
  {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.EventsTableID}}{{.BT}} events
WHERE
  event = 'issues'
  AND JSON_VALUE(payload, '$.repository.name') = 'breakglass'
  AND JSON_VALUE(payload, '$.organization.login') = '{{.Organization}}'
  AND JSON_VALUE(payload, '$.issue.user.login') = '{{.Author}}'
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.created_at')) <= TIMESTAMP('{{.Timestamp}}')
  AND TIMESTAMP(JSON_VALUE(payload, '$.issue.closed_at')) >= TIMESTAMP('{{.Timestamp}}')
  AND JSON_VALUE(payload, '$.issue.html_url') LIKE '{{.GitHubURLPrefix}}%'
`

type bgQueryParameters struct {
	ProjectID       string
	DatasetID       string
	EventsTableID   string
	Organization    string
	GitHubURLPrefix string
	Author          string
	Timestamp       string
	BT              string
}

// makeBreakglassQuery returns a BigQuery query that searches for a break glass
// issue created by given user and within a specified time frame.
func makeBreakglassQuery(cfg *Config, author, organization string, timestamp *time.Time) (string, error) {
	tmpl, err := template.New("breakglass-query").Parse(breakGlassIssueSQL)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &bgQueryParameters{
		ProjectID:       cfg.ProjectID,
		DatasetID:       cfg.DatasetID,
		EventsTableID:   cfg.EventsTableID,
		Organization:    organization,
		GitHubURLPrefix: ghURLPrefix(cfg.GitHub.GitHubEnterpriseServerURL),
		Author:          author,
		Timestamp:       timestamp.Format(time.RFC3339),
		BT:              "`",
	}); err != nil {
		return "", fmt.Errorf("failed to apply query template parameters: %w", err)
	}
	return sb.String(), nil
}
