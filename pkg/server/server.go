// Copyright 2022 GitHub Metrics Aggregator authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/abcxyz/pkg/logging"
	"github.com/verbanicm/gha-metrics/pkg/messaging"
	"go.uber.org/zap"
)

// Server provides the server implementation.
type Server struct {
	config   *Config
	messager messaging.Messager
	logger   *zap.SugaredLogger
}

// New creates a new instance of Server.
func New(ctx context.Context, config *Config, messager messaging.Messager) (*Server, error) {
	if messager == nil {
		return nil, fmt.Errorf("messager is required")
	}

	return &Server{
		config:   config,
		logger:   logging.FromContext(ctx),
		messager: messager,
	}, nil
}

// ServeHTTP creates the HTTP server and routes.
func (s *Server) ServeHTTP(ctx context.Context, handler *http.ServeMux) error {
	srv := &http.Server{Addr: ":" + s.config.Port, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		s.logger.Info("server.ServeHTTP: shutting down")
		errCh <- srv.Shutdown(shutdownCtx)
	}()

	s.logger.Infof("server.ServeHTTP: server listening at %v", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server.ServeHTTP: %v", err)
	}

	return nil
}
