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
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"

	"github.com/abcxyz/github-metrics-aggregator/pkg/webhook"
	"github.com/abcxyz/pkg/testutil"
)

func validateCfg(t *testing.T) *config {
	t.Helper()

	cfg, err := newTestConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return cfg
}

// createSignature creates a HMAC 256 signature for the test request payload.
func createSignature(key, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestHTTPEndpoints(t *testing.T) {
	t.Parallel()
	testutil.SkipIfNotIntegration(t)

	cfg := validateCfg(t)
	ctx := context.Background()
	requestID := uuid.New().String()

	resp, err := makeHTTPRequest(requestID, cfg.EndpointURL, cfg)
	if err != nil {
		t.Fatalf("error calling service url: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("invalid http response code got: %d, want: %d", resp.StatusCode, http.StatusCreated)
	}

	bqClient := makeBigQueryClient(ctx, t, cfg.ProjectID)
	query := generateQuery(bqClient, requestID, cfg.ProjectID, cfg.DatasetID)
	queryIfNumRowsExistWithRetries(ctx, t, query, cfg.QueryRetryWaitDuration, cfg.QueryRetryLimit, "test-main", 1)
}

func makeHTTPRequest(requestID, endpointURL string, cfg *config) (*http.Response, error) {
	payload, err := os.ReadFile(path.Join("..", "testdata", "pull_request.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to create payload from file: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpointURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log http request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.IDToken))
	req.Header.Add(webhook.DeliveryIDHeader, requestID)
	req.Header.Add(webhook.EventTypeHeader, "pull_request")
	req.Header.Add(webhook.SHA256SignatureHeader, fmt.Sprintf("sha256=%s", createSignature([]byte(cfg.GitHubWebhookSecret), payload)))

	httpClient := &http.Client{Timeout: cfg.HTTPRequestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute audit log request: %w", err)
	}
	return resp, nil
}

func generateQuery(bqClient *bigquery.Client, requestID, projectID, datasetID string) *bigquery.Query {
	queryString := fmt.Sprintf("SELECT COUNT(1) FROM `%s.%s`", projectID, datasetID)
	queryString += ` WHERE delivery_id = @request_id`
	return makeQuery(*bqClient, queryString, &[]bigquery.QueryParameter{{Name: "request_id", Value: requestID}})
}
