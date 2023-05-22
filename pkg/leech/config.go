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

package leech

import (
	"context"
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	BatchSize              int    `env:"BATCH_SIZE,default=100"`
	EventsProjectID        string `env:"EVENTS_PROJECT_ID"`
	EventsTable            string `env:"EVENTS_TABLE"`
	GitHubAppID            string `env:"GITHUB_APP_ID"`
	GitHubInstallID        string `env:"GITHUB_INSTALL_ID"`
	GitHubPrivateKey       string `env:"GITHUB_PRIVATE_KEY"`
	GitHubAppIDSecret      string `env:"GITHUB_APP_ID_SECRET"`
	GitHubInstallIDSecret  string `env:"GITHUB_INSTALL_ID_SECRET"`
	GitHubPrivateKeySecret string `env:"GITHUB_PRIVATE_KEY_SECRET"`
	JobName                string `env:"JOB_NAME"`
	LeechProjectID         string `env:"LEECH_PROJECT_ID"`
	LeechTable             string `env:"LEECH_TABLE"`
	LogsBucketName         string `env:"LOGS_BUCKET_NAME"`
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
	if cfg.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE must be a positive integer")
	}

	if cfg.EventsTable == "" {
		return fmt.Errorf("EVENTS_TABLE is required")
	}

	if cfg.EventsProjectID == "" {
		return fmt.Errorf("EVENTS_PROJECT_ID is required")
	}

	if cfg.GitHubAppID == "" && cfg.GitHubAppIDSecret == "" {
		return fmt.Errorf("GITHUB_APP_ID or GITHUB_APP_ID_SECRET is required")
	}

	if cfg.GitHubInstallID == "" && cfg.GitHubInstallIDSecret == "" {
		return fmt.Errorf("GITHUB_INSTALL_ID or GITHUB_INSTALL_ID_SECRET is required")
	}

	if cfg.GitHubPrivateKey == "" && cfg.GitHubPrivateKeySecret == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY or GITHUB_PRIVATE_KEY_SECRET is required")
	}

	if cfg.JobName == "" {
		return fmt.Errorf("JOB_NAME is required")
	}

	if cfg.LeechTable == "" {
		return fmt.Errorf("LEECH_TABLE is required")
	}

	if cfg.LeechProjectID == "" {
		return fmt.Errorf("LEECH_PROJECT_ID is required")
	}

	if cfg.LogsBucketName == "" {
		return fmt.Errorf("LOGS_BUCKET_NAME is required")
	}

	return nil
}

// ToFlags binds the config to the [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("COMMON PIPELINE OPTIONS")

	f.IntVar(&cli.IntVar{
		Name:   "batch-size",
		Target: &cfg.BatchSize,
		EnvVar: "BATCH_SIZE",
		Usage:  `The number of items to process in this pipeline run.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-project-id",
		Target: &cfg.EventsProjectID,
		EnvVar: "EVENTS_PROJECT_ID",
		Usage:  `The project id of the events table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-table",
		Target: &cfg.EventsTable,
		EnvVar: "EVENTS_TABLE",
		Usage:  `The dataset.table_name of the events table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id-secret",
		Target: &cfg.GitHubAppIDSecret,
		EnvVar: "GITHUB_APP_ID_SECRET",
		Usage:  `The secret name & version containing the GitHub App reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id",
		Target: &cfg.GitHubInstallID,
		EnvVar: "GITHUB_INSTALL_ID",
		Usage:  `The provisioned GitHub App Installation reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id-secret",
		Target: &cfg.GitHubInstallIDSecret,
		EnvVar: "GITHUB_INSTALL_ID_SECRET",
		Usage:  `The secret name & version containing the GitHub App installation reference.`,
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
		Name:   "job_name",
		Target: &cfg.JobName,
		EnvVar: "JOB_NAME",
		Usage:  `The beam job name provided by Dataflow.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "leech-project-id",
		Target: &cfg.LeechProjectID,
		EnvVar: "LEECH_PROJECT_ID",
		Usage:  `The project id of the leech table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "leech-table",
		Target: &cfg.LeechTable,
		EnvVar: "LEECH_TABLE",
		Usage:  `The dataset.table_name of the leech_status table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "logs-bucket-name",
		Target: &cfg.LogsBucketName,
		EnvVar: "LOGS_BUCKET_NAME",
		Usage:  `The name of the GCS bucket to store logs.`,
	})

	return set
}

// NewConfig creates a new Config from environment variables.
func NewConfig(ctx context.Context) (*Config, error) {
	return newConfig(ctx, envconfig.OsLookuper())
}

func newConfig(ctx context.Context, lu envconfig.Lookuper) (*Config, error) {
	var cfg Config
	if err := cfgloader.Load(ctx, &cfg, cfgloader.WithLookuper(lu)); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline config: %w", err)
	}
	if err := loadSecrets(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}
	return &cfg, nil
}

func loadSecrets(ctx context.Context, cfg *Config) error {
	// Load any secrets from secret manager
	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer sm.Close()
	if cfg.GitHubAppIDSecret != "" {
		res, err := readSecretText(ctx, sm, cfg.GitHubAppIDSecret)
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}
		cfg.GitHubAppID = res
	}
	if cfg.GitHubInstallIDSecret != "" {
		res, err := readSecretText(ctx, sm, cfg.GitHubInstallIDSecret)
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}
		cfg.GitHubInstallID = res
	}
	if cfg.GitHubPrivateKeySecret != "" {
		res, err := readSecretText(ctx, sm, cfg.GitHubPrivateKeySecret)
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}
		cfg.GitHubPrivateKeySecret = res
	}
	return nil
}

func readSecretText(ctx context.Context, client *secretmanager.Client, secretVersion string) (string, error) {
	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersion,
	}
	result, err := client.AccessSecretVersion(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("failed to get secret version for %q - %w", secretVersion, err)
	}
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("failed to get secret version for %q - data corrupted", secretVersion)
	}
	return string(result.Payload.Data), nil
}
