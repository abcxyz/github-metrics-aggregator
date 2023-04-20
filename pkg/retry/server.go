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

// Package retry is the retry server responsible for interacting with GitHub
// APIs to retry failed events.
package retry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/abcxyz/github-metrics-aggregator/pkg/clients"
	"github.com/abcxyz/pkg/healthcheck"
	"github.com/abcxyz/pkg/logging"
	"google.golang.org/api/option"
)

type Server struct {
	bqClient  *clients.BigQuery
	projectID string
}

// NewServer creates a new HTTP server implementation that will handle
// communication with GitHub APIs.
func NewServer(ctx context.Context, cfg *Config, bqClientOpts ...option.ClientOption) (*Server, error) {
	bq, err := clients.NewBigQuery(ctx, cfg.BigQueryProjectID, cfg.DatasetID, bqClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize BigQuery client: %w", err)
	}

	return &Server{
		bqClient:  bq,
		projectID: cfg.ProjectID,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes(ctx context.Context) http.Handler {
	logger := logging.FromContext(ctx)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthcheck.HandleHTTPHealthCheck())
	mux.Handle("/retry", s.handleRetry())

	// Middleware
	root := logging.HTTPInterceptor(logger, s.projectID)(mux)

	return root
}

// handleRetry handles calling GitHub APIs to search and retry for failed
// events.
func (s *Server) handleRetry() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `TODO: Make GitHub API calls and other stuff.. \n`)
	})
}
