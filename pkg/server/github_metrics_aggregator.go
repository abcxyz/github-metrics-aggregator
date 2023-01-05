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

// Package server is the base server for the github metrics ingestion.
package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/logging"
)

// GitHubMetricsAggregatorServer provides the server implementation.
type GitHubMetricsAggregatorServer struct {
	webhookSecret string
	messager      Messager
}

// Messager defines the functionality for sending messages.
type Messager interface {
	Send(ctx context.Context, msg []byte) error
}

// NewServer creates a new HTTP server implementation that will handle
// receiving webhook payloads.
func NewServer(ctx context.Context, webhookSecret string, messager Messager) (*GitHubMetricsAggregatorServer, error) {
	if messager == nil {
		return nil, fmt.Errorf("messager is required")
	}

	return &GitHubMetricsAggregatorServer{
		webhookSecret: webhookSecret,
		messager:      messager,
	}, nil
}

// handleWebhook creates a http.HandlerFunc implementation that processes webhook payload requests.
func (s *GitHubMetricsAggregatorServer) handleWebhook() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.FromContext(r.Context())

		respCode, respMsg, err := s.processWebhookRequest(r)
		if err != nil {
			logger.Errorw("error processing request", "code", respCode, "body", respMsg, "error", err)
		}
		w.WriteHeader(respCode)
		fmt.Fprint(w, respMsg)
	})
}

// handleVersion is a simple http.HandlerFunc that responds
// with version information for the server.
func (s *GitHubMetricsAggregatorServer) handleVersion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":%q}\n`, version.HumanVersion)
	})
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *GitHubMetricsAggregatorServer) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/webhook", s.handleWebhook())
	mux.Handle("/version", s.handleVersion())
	return mux
}
