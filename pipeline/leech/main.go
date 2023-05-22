// Copyright 2023 The Authors (see AUTHORS file)
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

// Package main contains a Beam data pipeline that will read workflow
// event records from BigQuery and ingest any available logs into cloud
// storage.
// The pipeline acts as a GitHub App for authentication purposes.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/cli"
	"github.com/abcxyz/github-metrics-aggregator/pkg/leech"
	"github.com/abcxyz/pkg/logging"
)

// init preregisters functions to speed up runtime reflection of types and function shapes.
func init() {
	leech.BeamInit()
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	logger := logging.FromContext(ctx)

	if err := realMain(ctx); err != nil {
		done()
		logger.Fatal(err)
		os.Exit(1)
	}
}

// realMain executes the Leech Pipeline.
func realMain(ctx context.Context) error {
	// Dataflow doesn't support commandline arguments that aren't in flag
	// format --flag=blah. Force the cli to execute the "leech pipeline"
	// subcommand.
	args := os.Args[1:]
	args = append([]string{"leech", "pipeline"}, args...)
	return cli.Run(ctx, args) //nolint:wrapcheck
}
