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

// Package main contains a Beam data pipeline that will read commit data
// from BigQuery, check each commits approval status, and write the approval
// status back to BigQuery.
package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/abcxyz/github-metrics-aggregator/pkg/auth"
	"github.com/abcxyz/github-metrics-aggregator/pkg/review"
	"github.com/abcxyz/github-metrics-aggregator/pkg/secrets"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	githubToken                     string
	githubAppID                     string
	githubAppInstallationID         string
	githubAppPrivateKeyResourceName string
	pushEventsTable                 string
	commitReviewStatusTable         string
	issuesTable                     string
)

func init() {
	// setup commandline arguments
	// explicitly *not* using the cli interface from abcxyz/pkg/cli due to conflicts
	// with Beam while using the Dataflow runner.
	flag.StringVar(&githubToken, "github-token", "", "The token to use to authenticate with github.")
	flag.StringVar(&githubAppID, "github-app-id", "", "The provisioned GitHub App reference.")
	flag.StringVar(&githubAppInstallationID, "github-app-installation-id", "", "The provisioned GitHub App Installation reference.")
	flag.StringVar(&githubAppPrivateKeyResourceName, "github-app-private-key-resource-name", "", "The resource name for the secret manager resource containing the GitHub App private key.")
	flag.StringVar(&pushEventsTable, "push-events-table", "", "The name of the push events table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&commitReviewStatusTable, "commit-review-status-table", "", "The of the commit review status table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flag.StringVar(&issuesTable, "issues-table", "", "The name of the issues table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		// beam/log is required in order for log severity to show up properly in
		// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
		// for more context.
		log.Errorf(ctx, "realMain failed: %v", err)
		os.Exit(1)
	}
}

// realMain executes the Commit Review Pipeline.
func realMain(ctx context.Context) error {
	// parse flags
	flag.Parse()
	// initialize beam. This must be done after flag.Parse() but before any flag validation.
	// https://cloud.google.com/dataflow/docs/guides/setting-pipeline-options#CreatePipelineFromArgs
	beam.Init()
	// parse commandline arguments into config
	pipelineConfig, err := getConfigFromFlags(ctx)
	if err != nil {
		return fmt.Errorf("failed to construct pipeline config: %w", err)
	}
	// construct and execute the pipeline
	pipeline := review.NewCommitApprovalPipeline(pipelineConfig)
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

// getConfigFromFlags returns review.CommitApprovalPipelineConfig struct
// using flag values. returns an error if any of the flags have malformed
// data.
func getConfigFromFlags(ctx context.Context) (*review.CommitApprovalPipelineConfig, error) {
	githubTokenSupplier, err := getGitHubTokenSupplier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct token supplier: %w", err)
	}
	token, err := githubTokenSupplier.GitHubToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %w", err)
	}
	qualifiedPushEventsTable, err := newQualifiedTableName(pushEventsTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pushEventsTable: %w", err)
	}
	qualifiedCommitReviewStatusTable, err := newQualifiedTableName(commitReviewStatusTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse commitReviewStatusTable: %w", err)
	}
	qualifiedIssuesTable, err := newQualifiedTableName(issuesTable)
	if err != nil {
		return nil, fmt.Errorf("unable to parse issuesTable: %w", err)
	}
	return &review.CommitApprovalPipelineConfig{
		GitHubAccessToken:       token,
		PushEventsTable:         *qualifiedPushEventsTable,
		CommitReviewStatusTable: *qualifiedCommitReviewStatusTable,
		IssuesTable:             *qualifiedIssuesTable,
	}, nil
}

// newQualifiedTableName parses a GoogleSQL table name, "<project>.<dataset>.<table>",
// into a bigqueryio.QualifiedTableName. This is in contrast to
// bigqueryio.NewQualifiedTableName, which parses a table name in BigQuery
// Legacy SQL format.
func newQualifiedTableName(s string) (*bigqueryio.QualifiedTableName, error) {
	c := strings.Index(s, ".")
	d := strings.LastIndex(s, ".")
	if c == -1 || d == -1 || d <= c {
		return nil, fmt.Errorf("table name missing components: %s", s)
	}

	project := s[:c]
	dataset := s[c+1 : d]
	table := s[d+1:]
	if strings.TrimSpace(project) == "" || strings.TrimSpace(dataset) == "" || strings.TrimSpace(table) == "" {
		return nil, fmt.Errorf("table name has empty components: %s", s)
	}
	return &bigqueryio.QualifiedTableName{Project: project, Dataset: dataset, Table: table}, nil
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
