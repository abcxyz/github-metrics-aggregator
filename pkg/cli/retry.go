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

	"github.com/sethvargo/go-gcslock"
	"google.golang.org/api/option"

	"github.com/abcxyz/github-metrics-aggregator/pkg/retry"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
)

var _ cli.Command = (*RetryJobCommand)(nil)

// The RetryJobCommand is a Cloud Run job that will read webhook events
// from BigQuery and retry any that have failed.
//
// The job acts as a GitHub App for authentication purposes.
type RetryJobCommand struct {
	cli.BaseCommand

	cfg *retry.Config

	// testFlagSetOpts is only used for testing.
	testFlagSetOpts []cli.Option

	// testDatastore is only used for testing
	testDatastore retry.Datastore

	// testGCSLockClientOptions is only used for testing
	testGCSLockClientOptions []option.ClientOption

	// testGCSLock is only used for testing
	testGCSLock gcslock.Lockable

	// testGitHub is only used for testing
	testGitHub retry.GitHubSource
}

func (c *RetryJobCommand) Desc() string {
	return `Execute a retry job for GitHub Metrics Aggregator`
}

func (c *RetryJobCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

	Execute a retry job for GitHub Metrics Aggregator.
`
}

func (c *RetryJobCommand) Flags() *cli.FlagSet {
	c.cfg = &retry.Config{}
	set := cli.NewFlagSet(c.testFlagSetOpts...)
	return c.cfg.ToFlags(set)
}

func (c *RetryJobCommand) Run(ctx context.Context, args []string) error {
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

	retryClientOptions := &retry.RetryClientOptions{}

	if c.testDatastore != nil {
		retryClientOptions.DatastoreClientOverride = c.testDatastore
	}

	if c.testGCSLockClientOptions != nil {
		retryClientOptions.GCSLockClientOpts = c.testGCSLockClientOptions
	}

	if c.testGCSLock != nil {
		retryClientOptions.GCSLockClientOverride = c.testGCSLock
	}

	if c.testGitHub != nil {
		retryClientOptions.GitHubOverride = c.testGitHub
	}

	if err := retry.ExecuteJob(ctx, c.cfg, retryClientOptions); err != nil {
		logger.ErrorContext(ctx, "error executing retry job", "error", err)
		return fmt.Errorf("job execution failed: %w", err)
	}

	return nil
}
