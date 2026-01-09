// Copyright 2025 The Authors (see AUTHORS file)
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

package relay

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"

	"github.com/abcxyz/github-metrics-aggregator/pkg/pubsub"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/logging"
)

// Server acts as a HTTP server for handling relay requests.
type Server struct {
	config          *Config
	relayMessenger  pubsub.Messenger
	messageEnricher MessageEnricher
}

// NewServer creates a new instance of the Server.
func NewServer(ctx context.Context, cfg *Config) (*Server, error) {
	agent := fmt.Sprintf("abcxyz:github-metrics-aggregator/relay/%s", version.Version)
	relayMessenger, err := pubsub.NewPubSubMessenger(ctx, cfg.RelayProjectID, cfg.RelayTopicID, option.WithUserAgent(agent))
	if err != nil {
		return nil, fmt.Errorf("failed to create event pubsub: %w", err)
	}
	messageEnricher := NewDefaultMessageEnricher()

	return &Server{
		config:          cfg,
		relayMessenger:  relayMessenger,
		messageEnricher: messageEnricher,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes(ctx context.Context) http.Handler {
	logger := logging.FromContext(ctx)
	mux := http.NewServeMux()
	mux.Handle("/", s.handleRelay())

	// Middleware
	root := logging.HTTPInterceptor(logger, s.config.ProjectID)(mux)

	return root
}
