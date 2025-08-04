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
	"fmt"
	"strings"

	"github.com/sethvargo/go-envconfig"

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the artifact job.
type Config struct {
	GitHubEnterpriseServerURL string `env:"GITHUB_ENTERPRISE_SERVER_URL"`  // The GitHub Enterprise Server instance URL, format "https://[hostname]"
	GitHubAppID               string `env:"GITHUB_APP_ID,required"`        // The GitHub App ID
	GitHubPrivateKeySecret    string `env:"GITHUB_PRIVATE_KEY_SECRET"`     // The GitHub App private key
	GitHubPrivateKeyKMSKeyID  string `env:"GITHUB_PRIVATE_KEY_KMS_KEY_ID"` // The KMS key ID of the GitHub App private key

	BatchSize int `env:"BATCH_SIZE,default=100"` // The number of items to process in this pipeline run

	ProjectID string `env:"PROJECT_ID,required"` // The project id where the tables live
	DatasetID string `env:"DATASET_ID,required"` // The dataset id where the tables live

	EventsTableID    string `env:"EVENTS_TABLE_ID,required"`    // The table_name of the events table
	ArtifactsTableID string `env:"ARTIFACTS_TABLE_ID,required"` // The table_name of the artifact_status table

	BucketName string `env:"BUCKET_NAME,required"` // The name of the GCS bucket to store artifact logs
}

// Validate validates the artifacts config after load.
func (cfg *Config) Validate() error {
	if cfg.GitHubEnterpriseServerURL != "" && !strings.HasPrefix(cfg.GitHubEnterpriseServerURL, "https://") {
		return fmt.Errorf("GITHUB_ENTERPRISE_SERVER_URL does not start with \"https://\"")
	}

	if cfg.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}

	if cfg.GitHubPrivateKeySecret == "" && cfg.GitHubPrivateKeyKMSKeyID == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY_SECRET or GITHUB_PRIVATE_KEY_KMS_KEY_ID is required")
	}

	if cfg.BucketName == "" {
		return fmt.Errorf("BUCKET_NAME is required")
	}

	if (cfg.EventsTableID) == "" {
		return fmt.Errorf("EVENTS_TABLE_ID is required")
	}

	if cfg.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
	}

	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID is required")
	}

	return nil
}

// NewConfig creates a new Config from environment variables.
func NewConfig(ctx context.Context) (*Config, error) {
	return newConfig(ctx, envconfig.OsLookuper())
}

func newConfig(ctx context.Context, lu envconfig.Lookuper) (*Config, error) {
	var cfg Config
	if err := cfgloader.Load(ctx, &cfg, cfgloader.WithLookuper(lu)); err != nil {
		return nil, fmt.Errorf("failed to parse retry server config: %w", err)
	}
	return &cfg, nil
}

// ToFlags binds the config to the [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("COMMON JOB OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:   "github-enterprise-server_url",
		Target: &cfg.GitHubEnterpriseServerURL,
		EnvVar: "GITHUB_ENTERPRISE_SERVER_URL",
		Usage:  `The GitHub Enterprise Server instance URL, format "http(s)://[hostname]"`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key-secret",
		Target: &cfg.GitHubPrivateKeySecret,
		EnvVar: "GITHUB_PRIVATE_KEY_SECRET",
		Usage:  `The GitHub App private key.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key-kms-key-id",
		Target: &cfg.GitHubPrivateKeySecret,
		EnvVar: "GITHUB_PRIVATE_KEY_KMS_KEY_ID",
		Usage:  `The KMS key ID of the GitHub App private key.`,
	})

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
