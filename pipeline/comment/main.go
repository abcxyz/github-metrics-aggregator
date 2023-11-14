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

// Package main contains a program that will read workflow event
// from BigQuery, comment on PRs with links to logs stored by leech, and
// store status back to BigQuery.
package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/auth"
	"github.com/abcxyz/github-metrics-aggregator/pkg/comment"
	"github.com/abcxyz/github-metrics-aggregator/pkg/secrets"
	"github.com/abcxyz/pkg/logging"
)

var (
	githubToken                     string
	githubAppID                     string
	githubAppInstallationID         string
	githubAppPrivateKeyResourceName string
	gcpProjectID                    string
	bigQueryDatasetID               string
)

const (
	defaultLogLevel = "info"
	defaultLogMode  = "development"
)

func init() {
	// setup commandline arguments
	// explicitly *not* using the cli interface from abcxyz/pkg/cli due to conflicts
	// with Beam while using the Dataflow runner.
	// TODO: since we are not using beam, may make sense to use abcxyz/pkg/cli, not sure
	flag.StringVar(&githubToken, "github-token", "", "The token to use to authenticate with github.") // TODO: should probalby use an env var instead for security. See my comment-test for an example.
	flag.StringVar(&githubAppID, "github-app-id", "", "The provisioned GitHub App reference.")
	flag.StringVar(&githubAppInstallationID, "github-app-installation-id", "", "The provisioned GitHub App Installation reference.")
	flag.StringVar(&githubAppPrivateKeyResourceName, "github-app-private-key-resource-name", "", "The resource name for the secret manager resource containing the GitHub App private key.")

	// Identifiers cannot be parameters in a query. I am hardcoding table name in sql.
	// Project and dataset id will instead be supplied by user.
	flag.StringVar(&gcpProjectID, "gcp-project-id", "", "The name of the GCP project that holds the GitHub Metrics Aggregator installation.")
	flag.StringVar(&gcpProjectID, "dataset-id", "github_metrics", "The name of the BigQuery dataset that holds the  GitHub Metrics Aggregator tables.")
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// realMain executes the Commit Review Pipeline.
func realMain(ctx context.Context) error {
	setLogEnvVars()
	ctx = logging.WithLogger(ctx, logging.NewFromEnv("GUARDIAN_"))
	// parse flags
	flag.Parse()
	// parse commandline arguments into config
	pipelineConfig, err := getConfigFromFlags(ctx)
	if err != nil {
		return fmt.Errorf("failed to construct pipeline config: %w", err)
	}

	return nil
}

// getConfigFromFlags returns comment.PrCommentPipelineConfig struct
// using flag values. returns an error if any of the flags have malformed
// data.
func getConfigFromFlags(ctx context.Context) (*comment.PrCommentPipelineConfig, error) {
	// TODO: use merr pattern from Seth's PR
	githubTokenSupplier, err := getGitHubTokenSupplier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct token supplier: %w", err)
	}
	token, err := githubTokenSupplier.GitHubToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %w", err)
	}
	return &comment.PrCommentPipelineConfig{
		GitHubAccessToken: token,
		GcpProjectID:      gcpProjectID,
		BigQueryDatasetID: bigQueryDatasetID,
	}, nil
}

// setLogEnvVars set the logging environment variables to their default
// values if not provided.
func setLogEnvVars() {
	if os.Getenv("ANALYZER_LOG_MODE") == "" {
		os.Setenv("ANALYZER_LOG_MODE", defaultLogMode)
	}

	if os.Getenv("ANALYZER_LOG_LEVEL") == "" {
		os.Setenv("ANALYZER_LOG_LEVEL", defaultLogLevel)
	}
}

func getGitHubTokenSupplier(ctx context.Context) (auth.GitHubTokenSupplier, error) {
	if githubToken != "" && githubAppID != "" {
		return nil, fmt.Errorf("both a githubToken and githubAppID were supplied")
	}
	if githubToken != "" {
		return auth.NewStaticGitHubTokenSupplier(githubToken), nil
	}
	if githubAppInstallationID == "" || githubAppPrivateKeyResourceName == "" {
		return nil, fmt.Errorf("both githubAppInstallationID and githubAppPrivateKeyResourceName must be supplied when using a githubAppID")
	}
	privateKey, err := getPrivateKey(ctx, githubAppPrivateKeyResourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}
	// TODO: seth's PR makes a lot of changes here.
	// TODO: I'm not certain which permissions you need to request to be able to comment
	// TODO: on a PR. My experiment in the comment-test works. I used a personal access token
	// TODO: with workflow, notifications, read:discussion and write:discussion permissions.
	// TODO: likely only a subset is needed.
	return auth.NewGitHubAppTokenSupplier(githubAppID, githubAppInstallationID, privateKey), nil
}

func getPrivateKey(ctx context.Context, secretResourceName string) (*rsa.PrivateKey, error) {
	privateKeyString, err := secrets.AccessSecretFromSecretManager(ctx, secretResourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key from secret manager: %w", err)
	}
	privateKey, err := secrets.ParsePrivateKey(privateKeyString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return privateKey, nil
}
