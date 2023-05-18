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
	"fmt"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/leech"
	"github.com/abcxyz/pkg/logging"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
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
		logger.Errorf("error processing pipeline: %w", err)
	}
}

// realMain executes the Leech Pipeline.
func realMain(ctx context.Context) error {
	// read configuration data
	cfg, err := leech.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("error reading configuration: %w", err)
	}
	// initialize beam and setup the pipeline
	beam.Init()
	p, s := beam.NewPipelineWithRoot()

	// create the main leech pipeline object and prepare it to run
	pipeline, err := leech.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error creating leech pipeline")
	}
	pipeline.Prepare(s)
	// execute the pipeline
	if err := beamx.Run(ctx, p); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}
