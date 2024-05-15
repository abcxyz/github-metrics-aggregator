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

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/sethvargo/go-envconfig"

	"github.com/abcxyz/github-metrics-aggregator/pkg/secrets"
	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the artifact job.
type Config struct {
	GitHubAppID            string `env:"GITHUB_APP_ID"`             // The GitHub App ID
	GitHubAppIDSecret      string `env:"GITHUB_APP_ID_SECRET"`      // The secret name & version containing the GitHub App ID
	GitHubInstallID        string `env:"GITHUB_INSTALL_ID"`         // The provisioned GitHub App Installation reference
	GitHubInstallIDSecret  string `env:"GITHUB_INSTALL_ID_SECRET"`  // The secret name & version containing the GitHub App installation ID
	GitHubPrivateKey       string `env:"GITHUB_PRIVATE_KEY"`        // The private key generated to call GitHub
	GitHubPrivateKeySecret string `env:"GITHUB_PRIVATE_KEY_SECRET"` // The secret name & version containing the GitHub App private key

	BatchSize int `env:"BATCH_SIZE,default=100"` // The number of items to process in this pipeline run

	ProjectID string `env:"PROJECT_ID,required"` // The project id where the tables live
	DatasetID string `env:"DATASET_ID,required"` // The dataset id where the tables live

	EventsTableID    string `env:"EVENTS_TABLE_ID,required"`    // The table_name of the events table
	ArtifactsTableID string `env:"ARTIFACTS_TABLE_ID,required"` // The table_name of the artifact_status table

	BucketName string `env:"BUCKET_NAME,required"` // The name of the GCS bucket to store artifact logs
}

// Validate validates the artifacts config after load.
func (cfg *Config) Validate() error {
	if cfg.GitHubAppID == "" && cfg.GitHubAppIDSecret == "" {
		return fmt.Errorf("one of [GITHUB_APP_ID | GITHUB_APP_ID_SECRET] is required")
	}
	if cfg.GitHubInstallID == "" && cfg.GitHubInstallIDSecret == "" {
		return fmt.Errorf("one of [GITHUB_INSTALL_ID | GITHUB_INSTALL_ID_SECRET] is required")
	}

	if cfg.GitHubPrivateKey == "" && cfg.GitHubPrivateKeySecret == "" {
		return fmt.Errorf("one of [GITHUB_PRIVATE_KEY | GITHUB_PRIVATE_KEY_SECRET] is required")
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
	f := set.NewSection("COMMON SERVER OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id-secret",
		Target: &cfg.GitHubAppIDSecret,
		EnvVar: "GITHUB_APP_ID_SECRET",
		Usage:  `The secret name & version containing the GitHub App ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_INSTALL_ID",
		Usage:  `The provisioned GitHub App installation ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id-secret",
		Target: &cfg.GitHubInstallIDSecret,
		EnvVar: "GITHUB_INSTALL_ID_SECRET",
		Usage:  `The secret name & version containing the GitHub App installation ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key",
		Target: &cfg.GitHubPrivateKey,
		EnvVar: "GITHUB_PRIVATE_KEY",
		Usage:  `The private key generated to call GitHub.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key-secret",
		Target: &cfg.GitHubPrivateKeySecret,
		EnvVar: "GITHUB_PRIVATE_KEY_SECRET",
		Usage:  `The secret name & version containing the GitHub App private key.`,
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

func (cfg *Config) ReplaceSecrets(ctx context.Context) error {
	// load any secrets from secret manager
	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer sm.Close()

	cfg.GitHubAppID, err = readSecretText(ctx, sm, cfg.GitHubAppIDSecret, cfg.GitHubAppID)
	if err != nil {
		return fmt.Errorf("failed to process app id secret: %w", err)
	}
	cfg.GitHubInstallID, err = readSecretText(ctx, sm, cfg.GitHubInstallIDSecret, cfg.GitHubInstallID)
	if err != nil {
		return fmt.Errorf("failed to process install id secret: %w", err)
	}
	cfg.GitHubPrivateKey, err = readSecretText(ctx, sm, cfg.GitHubPrivateKeySecret, cfg.GitHubPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to process private key secret: %w", err)
	}
	return nil
}

// readSecretText reads a value from Secret Manager if a secret version is provided, otherwise
// returns the defaultValue.
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
