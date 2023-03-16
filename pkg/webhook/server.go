// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// Server provides the server implementation.
type Server struct {
	webhookSecret string
	pubsub        *PubSubMessenger
}

// PubSubClientConfig are the pubsub client config options.
type PubSubClientConfig struct {
	PubSubURL      string
	PubSubGRPCConn *grpc.ClientConn
}

// NewServer creates a new HTTP server implementation that will handle
// receiving webhook payloads.
func NewServer(ctx context.Context, cfg *Config, pubsubClientOpts ...option.ClientOption) (*Server, error) {
	pubsub, err := NewPubSubMessenger(ctx, cfg.ProjectID, cfg.TopicID, pubsubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("server.NewPubSubMessenger: %w", err)
	}

	return &Server{
		webhookSecret: cfg.WebhookSecret,
		pubsub:        pubsub,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/webhook", s.handleWebhook())
	mux.Handle("/version", s.handleVersion())
	return mux
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
	return s.pubsub.Shutdown()
}
