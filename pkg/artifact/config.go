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

package artifact

import (
	"context"
	"errors"
	"fmt"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the artifact job.
type Config struct {
	GitHub githubclient.Config

	// BatchSize is the number of items to process in this pipeline run.
	BatchSize int

	// ProjectID is the project id where the tables live.
	ProjectID string

	// DatasetID is the dataset id where the tables live.
	DatasetID string

	// EventsTableID is the table_name of the events table.
	EventsTableID string

	// ArtifactsTableID is the table_name of the artifact_status table.
	ArtifactsTableID string

	// BucketName is the name of the GCS bucket to store artifact logs
	BucketName string
}

// Validate validates the artifacts config after load.
func (cfg *Config) Validate(ctx context.Context) error {
	var merr error

	merr = errors.Join(merr, cfg.GitHub.Validate(ctx))

	if cfg.BucketName == "" {
		merr = errors.Join(merr, fmt.Errorf("BUCKET_NAME is required"))
	}

	if (cfg.EventsTableID) == "" {
		merr = errors.Join(merr, fmt.Errorf("EVENTS_TABLE_ID is required"))
	}

	if cfg.ProjectID == "" {
		merr = errors.Join(merr, fmt.Errorf("PROJECT_ID is required"))
	}

	if cfg.DatasetID == "" {
		merr = errors.Join(merr, fmt.Errorf("DATASET_ID is required"))
	}

	return merr
}

// ToFlags binds the config to the [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	cfg.GitHub.ToFlags(set)

	f := set.NewSection("COMMON OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "bucket-name",
		Target:  &cfg.BucketName,
		EnvVar:  "BUCKET_NAME",
		Usage:   `The name of the bucket that holds artifact logs files from GitHub`,
		Example: "retry-lock-xxxx",
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-table-id",
		Target: &cfg.EventsTableID,
		EnvVar: "EVENTS_TABLE_ID",
		Usage:  `The events table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "artifacts-table-id",
		Target: &cfg.ArtifactsTableID,
		EnvVar: "ARTIFACTS_TABLE_ID",
		Usage:  `The artifacts table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "dataset-id",
		Target: &cfg.DatasetID,
		EnvVar: "DATASET_ID",
		Usage:  `BigQuery dataset ID.`,
	})

	f.IntVar(&cli.IntVar{
		Name:    "batch-size",
		Target:  &cfg.BatchSize,
		EnvVar:  "BATCH_SIZE",
		Default: 100,
		Usage:   `The number of items to process in this execution`,
	})

	return set
}
