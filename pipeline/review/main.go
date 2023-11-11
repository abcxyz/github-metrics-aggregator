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
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/abcxyz/github-metrics-aggregator/pkg/githubauth"
	"github.com/abcxyz/github-metrics-aggregator/pkg/review"
	"github.com/abcxyz/github-metrics-aggregator/pkg/secrets"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
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
	flagGitHubToken := flag.String("github-token", "",
		"Token to use to authenticate with github.")
	flagGitHubAppID := flag.String("github-app-id", "",
		"Provisioned GitHub App reference.")
	flagGitHubAppInstallationID := flag.String("github-app-installation-id", "",
		"Provisioned GitHub App Installation reference.")
	flagGitHubAppPrivateKeyResourceName := flag.String("github-app-private-key-resource-name", "",
		"Resource name for the secret manager resource containing the GitHub App private key.")
	flagPushEventsTable := flag.String("push-events-table", "",
		"Name of the push events table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flagCommitReviewStatusTable := flag.String("commit-review-status-table", "",
		"Name of the commit review status table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")
	flagIssuesTable := flag.String("issues-table", "",
		"name of the issues table. The value provided must be a fully qualified BigQuery table name of the form <project>:<dataset>.<table>")

	// Parse flags.
	flag.Parse()

	// Initialize beam. This must be done after flag.Parse(), but before any flag
	// validation.
	// https://cloud.google.com/dataflow/docs/guides/setting-pipeline-options#CreatePipelineFromArgs
	beam.Init()

	// Validate inputs.
	var merr error
	if *flagGitHubToken == "" && *flagGitHubAppID == "" {
		merr = errors.Join(merr, fmt.Errorf("one of github token or github app id are required"))
	}
	if *flagGitHubToken != "" && *flagGitHubAppID != "" {
		merr = errors.Join(merr, fmt.Errorf("only one of github token or github app id are allowed"))
	}

	// Create qualified table names.
	pushEventsTable, err := newQualifiedTableName(*flagPushEventsTable)
	merr = errors.Join(merr, err)
	commitReviewStatusTable, err := newQualifiedTableName(*flagCommitReviewStatusTable)
	merr = errors.Join(merr, err)
	issuesTable, err := newQualifiedTableName(*flagIssuesTable)
	merr = errors.Join(merr, err)

	// Create the GitHub token source.
	var githubTokenSource githubauth.TokenSource
	if *flagGitHubToken != "" {
		githubTokenSource, err = githubauth.NewStaticTokenSource(*flagGitHubToken)
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("failed to create github static token source: %w", err))
			return merr
		}
	} else {
		sm, err := secretmanager.NewClient(ctx)
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("failed to create secret manager client: %w", err))
			return merr
		}
		defer sm.Close()

		privateKeyPEM, err := secrets.AccessSecret(ctx, sm, *flagGitHubAppPrivateKeyResourceName)
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("failed to get github private key pem: %w", err))
			return merr
		}

		githubTokenSource, err = githubauth.NewAppTokenSource(*flagGitHubAppID, *flagGitHubAppInstallationID, privateKeyPEM, githubauth.ForAllRepos())
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("failed to create github app token source: %w", err))
			return merr
		}
	}
	githubToken, err := githubTokenSource.GitHubToken(ctx)
	if err != nil {
		merr = errors.Join(merr, fmt.Errorf("failed to get github token: %w", err))
		return merr
	}

	if merr != nil {
		return merr
	}

	// Construct and execute the pipeline.
	pipeline := review.NewCommitApprovalPipeline(&review.CommitApprovalPipelineConfig{
		GitHubToken:             githubToken,
		PushEventsTable:         pushEventsTable,
		CommitReviewStatusTable: commitReviewStatusTable,
		IssuesTable:             issuesTable,
	})
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

// newQualifiedTableName parses a GoogleSQL table name,
// "<project>.<dataset>.<table>", into a bigqueryio.QualifiedTableName.
//
// This is in contrast to bigqueryio.NewQualifiedTableName, which parses a table
// name in BigQuery Legacy SQL format.
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
