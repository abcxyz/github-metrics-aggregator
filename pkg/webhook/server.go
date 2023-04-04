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

// Package webhook is the base webhook server for the github metrics ingestion.
package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/abcxyz/github-metrics-aggregator/pkg/clients"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/healthcheck"
	"github.com/abcxyz/pkg/logging"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// Server provides the server implementation.
type Server struct {
	eventsTableID       string
	failureEventTableID string
	eventPubsub         *clients.PubSubMessenger
	dlqPubsub           *clients.PubSubMessenger
	retryLimit          int
	webhookSecret       string
	projectID           string
}

// PubSubClientConfig are the pubsub client config options.
type PubSubClientConfig struct {
	PubSubURL      string
	PubSubGRPCConn *grpc.ClientConn
}

// NewServer creates a new HTTP server implementation that will handle
// receiving webhook payloads.
func NewServer(ctx context.Context, cfg *Config, pubsubClientOpts ...option.ClientOption) (*Server, error) {
	eventPubsub, err := clients.NewPubSubMessenger(ctx, cfg.ProjectID, cfg.EventTopicID, pubsubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create event pubsub: %w", err)
	}

	dlqPubsub, err := clients.NewPubSubMessenger(ctx, cfg.ProjectID, cfg.DLQEventTopicID, pubsubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ pubsub: %w", err)
	}

	return &Server{
		eventsTableID:       cfg.EventsTableID,
		failureEventTableID: cfg.FailureEventsTableID,
		projectID:           cfg.ProjectID,
		eventPubsub:         eventPubsub,
		dlqPubsub:           dlqPubsub,
		retryLimit:          cfg.RetryLimit,
		webhookSecret:       cfg.WebhookSecret,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes(ctx context.Context) http.Handler {
	logger := logging.FromContext(ctx)
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthcheck.HandleHTTPHealthCheck())
	mux.Handle("/webhook", s.handleWebhook())
	mux.Handle("/version", s.handleVersion())

	// Middleware
	root := logging.HTTPInterceptor(logger, s.projectID)(mux)

	return root
}

// handleVersion is a simple http.HandlerFunc that responds
// with version information for the server.
func (s *Server) handleVersion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":%q}\n`, version.HumanVersion)
	})
}

// Shutdown handles the graceful shutdown of the webhook server.
func (s *Server) Shutdown() error {
	if err := s.eventPubsub.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown event pubsub connection: %w", err)
	}

	if err := s.dlqPubsub.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown DLQ pubsub connection: %w", err)
	}

	return nil
}
