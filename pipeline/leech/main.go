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

// Package main contains a Beam data pipeline that will read workflow event
// records from BigQuery and ingest any available logs into cloud storage.
//
// The pipeline acts as a GitHub App for authentication purposes.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"

	"github.com/abcxyz/github-metrics-aggregator/pkg/leech"
	"github.com/abcxyz/github-metrics-aggregator/pkg/secrets"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/logging"
)

// init preregisters functions to speed up runtime reflection of types and function shapes.
func init() {
	register.DoFn2x1[context.Context, leech.EventRecord, leech.LeechRecord](&leech.IngestLogsFn{})
}

// main is the pipeline entry point called by the beam runner.
func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	logger := logging.FromContext(ctx)

	if err := realMain(ctx); err != nil {
		done()
		logger.ErrorContext(ctx, "process exited with error", err)
		os.Exit(1)
	}
}

// realMain executes the Leech Pipeline.
func realMain(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	// setup commandline arguments
	// explicitly *not* using the cli interface from abcxyz/pkg/cli due to conflicts
	// with Beam while using the Dataflow runner.
	batchSize := flag.Int("batch-size", 100, "The number of items to process in this pipeline run.")
	eventsProjectID := flag.String("events-project-id", "", "The project id of the events table.")
	eventsTable := flag.String("events-table", "", "The dataset.table_name of the events table.")
	githubAppID := flag.String("github-app-id", "", "The provisioned GitHub App reference.")
	githubAppIDSecret := flag.String("github-app-id-secret", "", "The secret name & version containing the GitHub App reference.")
	githubInstallID := flag.String("github-install-id", "", "The provisioned GitHub App Installation reference.")
	githubInstallIDSecret := flag.String("github-install-id-secret", "", "The secret name & version containing the GitHub App installation reference.")
	githubPrivateKey := flag.String("github-private-key", "", "The private key generated to call GitHub.")
	githubPrivateKeySecret := flag.String("github-private-key-secret", "", "The secret name & version containing the GitHub App private key.")
	leechProjectID := flag.String("leech-project-id", "", "The project id of the leech table.")
	leechTable := flag.String("leech-table", "", "The dataset.table_name of the leech_status table.")
	logsBucketName := flag.String("logs-bucket-name", "", "The name of the GCS bucket to store logs.")

	flag.Parse()

	// initialize beam
	beam.Init()

	// setup the beam pipeline
	pipeline, scope := beam.NewPipelineWithRoot()

	// BigQuery table notation is not consistent so it needs represented in a few different
	// formats to appease the BigQuery client.
	eventsTableDotNotation := fmt.Sprintf("%s.%s", *eventsProjectID, *eventsTable)
	leechTableDotNotation := fmt.Sprintf("%s.%s", *leechProjectID, *leechTable)
	leechTableColonNotation := fmt.Sprintf("%s:%s", *leechProjectID, *leechTable)

	// load any secrets from secret manager
	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer sm.Close()

	appID, err := readSecretText(ctx, sm, *githubAppIDSecret, *githubAppID)
	if err != nil {
		return fmt.Errorf("failed to process app id secret: %w", err)
	}
	installID, err := readSecretText(ctx, sm, *githubInstallIDSecret, *githubInstallID)
	if err != nil {
		return fmt.Errorf("failed to process install id secret: %w", err)
	}
	privateKey, err := readSecretText(ctx, sm, *githubPrivateKeySecret, *githubPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to process private key secret: %w", err)
	}

	var event leech.EventRecord
	query := fmt.Sprintf(leech.SourceQuery, eventsTableDotNotation, leechTableDotNotation, *batchSize)
	// step 1: query BigQuery for unprocessed events
	col := bigqueryio.Query(scope, *leechProjectID, query, reflect.TypeOf(event), bigqueryio.UseStandardSQL())
	// step 2: process the events in parallel, ingesting logs
	res := beam.ParDo(scope, &leech.IngestLogsFn{
		LogsBucketName:   *logsBucketName,
		GitHubAppID:      appID,
		GitHubInstallID:  installID,
		GitHubPrivateKey: privateKey,
	}, col)
	// step 3: write all of the results back to BigQuery
	bigqueryio.Write(scope, *leechProjectID, leechTableColonNotation, res)

	logger.InfoContext(ctx, "pipeline starting",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	// execute the pipeline
	if err := beamx.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

func readSecretText(ctx context.Context, client *secretmanager.Client, secretVersion, defaultValue string) (string, error) {
	// if the secret version is empty fallback on the default value
	if secretVersion == "" {
		return defaultValue, nil
	}
	secret, err := secrets.AccessSecret(ctx, client, secretVersion)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}
	return secret, nil
}
