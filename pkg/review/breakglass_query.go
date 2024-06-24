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

// breakGlassIssueSQL is the BigQuery query that searches for a
// break glass issues created by given user and within a specified time frame.
const breakGlassIssueSQL = `
SELECT
  issues.html_url html_url
FROM
  {{.BT}}{{.ProjectID}}.{{.DatasetID}}.{{.IssuesTableID}}{{.BT}} issues
WHERE
  issues.repository = 'breakglass'
  AND author = '{{.Author}}'
  AND issues.created_at <= TIMESTAMP('{{.Timestamp}}')
  AND issues.closed_at >= TIMESTAMP('{{.Timestamp}}')
`

type bgQueryParameters struct {
	ProjectID     string
	DatasetID     string
	IssuesTableID string
	Author        string
	Timestamp     string
	BT            string
}

// makeBreakglassQuery returns a BigQuery query that searches for a break glass
// issue created by given user and within a specified time frame.
func makeBreakglassQuery(cfg *Config, author, timestamp string) (string, error) {
	tmpl, err := template.New("breakglass-query").Parse(breakGlassIssueSQL)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, &bgQueryParameters{
		ProjectID:     cfg.ProjectID,
		DatasetID:     cfg.DatasetID,
		IssuesTableID: cfg.IssuesTableID,
		Author:        author,
		Timestamp:     timestamp,
		BT:            "`",
	}); err != nil {
		return "", fmt.Errorf("failed to apply query template parameters: %w", err)
	}
	return sb.String(), nil
}
