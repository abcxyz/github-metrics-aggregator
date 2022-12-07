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

// Package main is the main entrypoint to the application
package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/abcxyz/pkg/logging"
	_ "github.com/joho/godotenv/autoload"
	"github.com/verbanicm/gha-metrics/pkg/messaging"
	"github.com/verbanicm/gha-metrics/pkg/server"
)

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

func realMain(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	config, err := server.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("server.Setup: %w", err)
	}

	pubsubMessager, err := messaging.NewPubSubMessager(ctx, config.ProjectID, config.TopicID, logger)
	if err != nil {
		return fmt.Errorf("messaging.NewPubSubMessager: %w", err)
	}
	defer pubsubMessager.Cleanup(ctx)

	srv, err := server.New(ctx, config, pubsubMessager)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/webhook", srv.HandleWebhook)

	if err := srv.ServeHTTP(ctx, handler); err != nil {
		return fmt.Errorf("server.ServeHTTP: %w", err)
	}

	return nil
}
