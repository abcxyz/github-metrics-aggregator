// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sethvargo/go-retry"
)

// makeBigQueryClient creates a new client and automatically closes the
// connection when the tests finish.
func makeBigQueryClient(ctx context.Context, tb testing.TB, projectID string) *bigquery.Client {
	tb.Helper()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := client.Close(); err != nil {
			tb.Errorf("failed to close the biquery client: %v", err)
		}
	})

	return client
}

// makeQuery generates a bigquery Query from a query string and parameters.
func makeQuery(bqClient bigquery.Client, queryString string, queryParameters *[]bigquery.QueryParameter) *bigquery.Query {
	bqQuery := bqClient.Query(queryString)
	bqQuery.Parameters = *queryParameters
	return bqQuery
}

// queryIfNumRowsExistWithRetries queries the DB and checks if the expected number of rows exists or not.
// If not, the query will be retried with the specified retry inputs.
func queryIfNumRowsExistWithRetries(ctx context.Context, tb testing.TB, bqQuery *bigquery.Query, retryWaitDuration time.Duration, retryLimit uint64, testName string, expectedNum int64) {
	tb.Helper()

	b := retry.NewExponential(retryWaitDuration)
	if err := retry.Do(ctx, retry.WithMaxRetries(retryLimit, b), func(ctx context.Context) error {
		found, err := queryIfNumRowsExist(ctx, tb, bqQuery, expectedNum)
		if found {
			// Early exit retry if rows already found.
			return nil
		}

		tb.Log("matching entry not found, retrying...")

		if err != nil {
			tb.Logf("query error: %v.", err)
		}
		return retry.RetryableError(fmt.Errorf("no matching records found in bigquery after timeout for %q", testName))
	}); err != nil {
		tb.Errorf("Retry failed: %v.", err)
	}
}

// queryIfNumRowsExist queries the DB and checks if the expected number of rows exists or not.
func queryIfNumRowsExist(ctx context.Context, tb testing.TB, query *bigquery.Query, expectedNum int64) (bool, error) {
	tb.Helper()

	q, err := query.Run(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to run query: %w", err)
	}

	it, err := q.Read(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to read query results: %w", err)
	}

	var row []bigquery.Value
	if err := it.Next(&row); err != nil {
		return false, fmt.Errorf("failed to get next row: %w", err)
	}

	// Check if the matching row count is equal to expected, if yes, then the rows exists.
	tb.Logf("Found %d matching rows", row[0])
	result, ok := row[0].(int64)
	if !ok {
		return false, fmt.Errorf("error converting query results to integer value (got %T)", row[0])
	}
	return result == expectedNum, nil
}
