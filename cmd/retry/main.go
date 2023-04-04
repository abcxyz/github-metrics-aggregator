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

package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/retry"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/serving"
)

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

// realMain creates an HTTP server meant to called by Cloud Scheduler.
func realMain(ctx context.Context) error {
	cfg, err := retry.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	retryServer, err := retry.NewServer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	server, err := serving.New(cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to create serving infrastructure: %w", err)
	}
	return server.StartHTTPHandler(ctx, retryServer.Routes())
}
