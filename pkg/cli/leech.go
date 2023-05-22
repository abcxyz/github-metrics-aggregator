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

package cli

import (
	"context"
	"fmt"

	"github.com/abcxyz/github-metrics-aggregator/pkg/leech"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var _ cli.Command = (*LeechCommand)(nil)

type LeechCommand struct {
	cli.BaseCommand

	cfg *leech.Config

	// testFlagSetOpts is only used for testing.
	testFlagSetOpts []cli.Option
}

func (c *LeechCommand) Desc() string {
	return `Start a leech pipeline for GitHub Metrics Aggregator`
}

func (c *LeechCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]
  Start a leech pipeline for GitHub Metrics Aggregator.
`
}

func (c *LeechCommand) Flags() *cli.FlagSet {
	c.cfg = &leech.Config{}
	set := cli.NewFlagSet(c.testFlagSetOpts...)
	return c.cfg.ToFlags(set)
}

func (c *LeechCommand) Run(ctx context.Context, args []string) error {
	store, err := leech.NewObjectStore(ctx)
	if err != nil {
		return fmt.Errorf("failed to create object store client: %w", err)
	}

	pipeline, err := c.RunUnstarted(ctx, args, store)
	if err != nil {
		return err
	}

	// execute the pipeline
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

func (c *LeechCommand) RunUnstarted(ctx context.Context, args []string, storage leech.ObjectWriter) (*beam.Pipeline, error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %q", args)
	}

	logger := logging.FromContext(ctx)
	logger.Debugw("pipeline starting",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	if err := c.cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	logger.Debugw("loaded configuration", "config", c.cfg)

	// initialize beam and setup the pipeline
	beam.Init()
	p, s := beam.NewPipelineWithRoot()

	// create the main leech pipeline object and prepare it to run
	pipeline, err := leech.New(ctx, c.cfg, storage)
	if err != nil {
		return nil, fmt.Errorf("error creating leech pipeline: %w", err)
	}
	pipeline.Prepare(s)
	return p, nil
}
