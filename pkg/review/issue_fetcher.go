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
	"context"
	"fmt"
	"time"

	"github.com/abcxyz/github-metrics-aggregator/pkg/bq"
)

// BreakGlassIssueFetcher fetches break glass issues from a data source.
type BreakGlassIssueFetcher interface {
	// getBreakGlassIssues retrieves all break glass issues created by the given
	// author and whose open duration contains the specified timestamp.
	// The issue's open duration contains the timestamp if
	// issue.created_at <= timestamp <= issue.closed_at holds.
	fetch(ctx context.Context, cfg *Config, author string, timestamp time.Time) ([]*breakGlassIssue, error)
}

// BigQueryBreakGlassIssueFetcher implements the BreakGlassIssueFetcher
// interface and fetches the break glass issue data from BigQuery.
type BigQueryBreakGlassIssueFetcher struct {
	client *bq.BigQuery
}

func (bqif *BigQueryBreakGlassIssueFetcher) fetch(ctx context.Context, cfg *Config, author string, timestamp time.Time) ([]*breakGlassIssue, error) {
	issueQuery, err := makeBreakglassQuery(cfg, author, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create breakglass query: %w", err)
	}
	items, err := bq.Query[breakGlassIssue](ctx, bqif.client, issueQuery)
	if err != nil {
		return nil, fmt.Errorf("client.Query failed: %w", err)
	}
	return items, nil
}
