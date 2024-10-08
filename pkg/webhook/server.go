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

	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/healthcheck"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/renderer"
)

// Datastore adheres to the interaction the webhook service has with a datastore.
type Datastore interface {
	DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error)
	FailureEventsExceedsRetryLimit(ctx context.Context, failureEventTableID, deliveryID string, retryLimit int) (bool, error)
	WriteFailureEvent(ctx context.Context, failureEventTableID, deliveryID, createdAt string) error
	Close() error
}

// Server provides the server implementation.
type Server struct {
	h                   *renderer.Renderer
	datastore           Datastore
	eventsTableID       string
	failureEventTableID string
	eventsPubsub        *PubSubMessenger
	dlqEventsPubsub     *PubSubMessenger
	retryLimit          int
	webhookSecret       string
	projectID           string
}

// PubSubClientConfig are the pubsub client config options.
type PubSubClientConfig struct {
	PubSubURL      string
	PubSubGRPCConn *grpc.ClientConn
}

// WebhookClientOptions encapsulate client config options as well as dependency implementation overrides.
type WebhookClientOptions struct {
	EventPubsubClientOpts    []option.ClientOption
	DLQEventPubsubClientOpts []option.ClientOption
	BigQueryClientOpts       []option.ClientOption
	DatastoreClientOverride  Datastore // used for unit testing
}

// NewServer creates a new HTTP server implementation that will handle
// receiving webhook payloads.
func NewServer(ctx context.Context, h *renderer.Renderer, cfg *Config, wco *WebhookClientOptions) (*Server, error) {
	eventsPubsub, err := NewPubSubMessenger(ctx, cfg.ProjectID, cfg.EventsTopicID, wco.EventPubsubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create event pubsub: %w", err)
	}

	dlqEventsPubsub, err := NewPubSubMessenger(ctx, cfg.ProjectID, cfg.DLQEventsTopicID, wco.DLQEventPubsubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ pubsub: %w", err)
	}

	datastore := wco.DatastoreClientOverride
	if datastore == nil {
		bq, err := NewBigQuery(ctx, cfg.BigQueryProjectID, cfg.DatasetID, wco.BigQueryClientOpts...)
		if err != nil {
			return nil, fmt.Errorf("server.NewBigQuery: %w", err)
		}
		datastore = bq
	}

	return &Server{
		h:                   h,
		datastore:           datastore,
		eventsTableID:       cfg.EventsTableID,
		failureEventTableID: cfg.FailureEventsTableID,
		eventsPubsub:        eventsPubsub,
		dlqEventsPubsub:     dlqEventsPubsub,
		projectID:           cfg.ProjectID,
		retryLimit:          cfg.RetryLimit,
		webhookSecret:       cfg.GitHubWebhookSecret,
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

// handleVersion is a simple http.HandlerFunc that responds with version
// information for the server.
func (s *Server) handleVersion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.h.RenderJSON(w, http.StatusOK, map[string]string{
			"version": version.HumanVersion,
		})
	})
}

// Close handles the graceful shutdown of the webhook server.
func (s *Server) Close() error {
	if err := s.eventsPubsub.Close(); err != nil {
		return fmt.Errorf("failed to shutdown event pubsub connection: %w", err)
	}

	if err := s.dlqEventsPubsub.Close(); err != nil {
		return fmt.Errorf("failed to shutdown DLQ pubsub connection: %w", err)
	}

	if err := s.datastore.Close(); err != nil {
		return fmt.Errorf("failed to close the BigQuery connection: %w", err)
	}

	return nil
}
