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

// Package main is the main entrypoint to the application.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/abcxyz/github-metrics-aggregator/pkg/server"
	"github.com/abcxyz/pkg/logging"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

const userAgent = "abcxyz/github-metrics-aggregator"

// main is the application entry point. It primarily wraps the realMain function with
// a context that properly handles signals from the OS.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	logger := logging.NewFromEnv("")
	ctx = logging.WithLogger(ctx, logger)

	if err := realMain(ctx); err != nil {
		done()
		logger.Fatal(err)
	}
}

// realMain creates an HTTP server to receive GitHub webhook payloads
// This server supports graceful stopping and cancellation by:
//   - using a cancellable context
//   - listening to incoming requests in a goroutine
func realMain(ctx context.Context) error {
	cfg, err := server.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("server.NewConfig: %w", err)
	}

	pubsubClientOpts := []option.ClientOption{option.WithUserAgent(userAgent)}
	webhookServer, err := server.NewServer(ctx, cfg, pubsubClientOpts...)
	if err != nil {
		return fmt.Errorf("server.NewServer: %w", err)
	}

	// Create the server and listen in a goroutine.
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           webhookServer.Routes(),
		ReadHeaderTimeout: 2 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case serverErrCh <- err:
			default:
			}
		}
	}()

	// Wait for shutdown signal or error from the listener.
	select {
	case err := <-serverErrCh:
		return fmt.Errorf("error from server listener: %w", err)
	case <-ctx.Done():
	}

	// Gracefully shut down the server.
	shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	if err := webhookServer.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup webhook server: %w", err)
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
