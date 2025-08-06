// Copyright 2024 The Authors (see AUTHORS file)
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

package cli

import (
	"context"
	"fmt"

	"github.com/abcxyz/github-metrics-aggregator/pkg/review"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
)

var _ cli.Command = (*ReviewJobCommand)(nil)

// The ReviewJobCommand is a Cloud Run job that will read commits
// from BigQuery and verify that they received proper review.
//
// The job acts as a GitHub App for authentication purposes.
type ReviewJobCommand struct {
	cli.BaseCommand

	cfg *review.Config

	// testFlagSetOpts is only used for testing.
	testFlagSetOpts []cli.Option
}

func (c *ReviewJobCommand) Desc() string {
	return `Execute an artifact retrieval job for GitHub Metrics Aggregator`
}

func (c *ReviewJobCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]
	Execute an artifact retrieval job for GitHub Metrics Aggregator
`
}

func (c *ReviewJobCommand) Flags() *cli.FlagSet {
	c.cfg = &review.Config{}
	set := cli.NewFlagSet(c.testFlagSetOpts...)
	return c.cfg.ToFlags(set)
}

func (c *ReviewJobCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %q", args)
	}

	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "running job",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	if err := c.cfg.Validate(ctx); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	logger.DebugContext(ctx, "loaded configuration", "config", c.cfg)

	if err := review.ExecuteJob(ctx, c.cfg); err != nil {
		return fmt.Errorf("job execution failed: %w", err)
	}

	return nil
}
