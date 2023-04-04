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
	"fmt"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/webhook"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/serving"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

const userAgent = "abcxyz/github-metrics-aggregator"

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	logger := logging.NewFromEnv("")
	ctx = logging.WithLogger(ctx, logger)

	if err := realMain(ctx); err != nil {
		done()
		logger.Fatal(err)
	}
}

// realMain creates an HTTP server to receive GitHub webhook payloads.
func realMain(ctx context.Context) error {
	cfg, err := webhook.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	pubsubClientOpts := []option.ClientOption{option.WithUserAgent(userAgent)}
	webhookServer, err := webhook.NewServer(ctx, cfg, pubsubClientOpts...)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	server, err := serving.New(cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to create serving infrastructure: %w", err)
	}
	return server.StartHTTPHandler(ctx, webhookServer.Routes())
}
