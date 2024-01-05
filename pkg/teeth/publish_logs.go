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

// Package teeth contains a job that will read leech and pull request event
// records from BigQuery and publish any available log invocations in a PR
// comment.
package teeth

import (
	"context"
	"fmt"
)

// BigQueryClient defines the spec for calls to read from and write to
// BigQuery tables.
type BigQueryClient interface {
	QueryLatest(context.Context) ([]*PublisherSourceRecord, error)
	Insert(context.Context, []*InvocationCommentStatusRecord) error
}

// GetLatestSourceRecords gets the latest publisher source records.
func GetLatestSourceRecords(ctx context.Context, bqClient BigQueryClient) ([]*PublisherSourceRecord, error) {
	res, err := bqClient.QueryLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query with bigquery client: %w", err)
	}
	return res, nil
}

// SaveInvocationCommentStatus inserts the statuses into the
// InvocationCommentStatus table.
func SaveInvocationCommentStatus(ctx context.Context, bqClient BigQueryClient, statuses []*InvocationCommentStatusRecord) error {
	if err := bqClient.Insert(ctx, statuses); err != nil {
		return fmt.Errorf("failed to insert with bigquery client: %w", err)
	}
	return nil
}
