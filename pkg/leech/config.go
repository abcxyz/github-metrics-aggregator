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

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	BatchSize        int    `env:"BATCH_SIZE,required,default=100"`
	EventsProjectID  string `env:"EVENTS_PROJECT_ID,required"`
	EventsTable      string `env:"EVENTS_TABLE,required"`
	GitHubAppID      string `env:"GITHUB_APP_ID,required"`
	GitHubInstallID  string `env:"GITHUB_INSTALL_ID,required"`
	GitHubPrivateKey string `env:"GITHUB_PRIVATE_KEY,required"`
	LeechProjectID   string `env:"LEECH_PROJECT_ID,required"`
	LeechTable       string `env:"LEECH_TABLE,required"`
	LogsBucketName   string `env:"LOGS_BUCKET_NAME,required"`
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
	if cfg.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE is required to be a positive integer")
	}

	if cfg.EventsTable == "" {
		return fmt.Errorf("EVENTS_TABLE is required")
	}

	if cfg.EventsProjectID == "" {
		return fmt.Errorf("EVENTS_PROJECT_ID is required")
	}

	if cfg.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}

	if cfg.GitHubInstallID == "" {
		return fmt.Errorf("GITHUB_INSTALL_ID is required")
	}

	if cfg.GitHubPrivateKey == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY is required")
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

// NewConfig creates a new Config from environment variables.
func NewConfig(ctx context.Context) (*Config, error) {
	return newConfig(ctx, envconfig.OsLookuper())
}

func newConfig(ctx context.Context, lu envconfig.Lookuper) (*Config, error) {
	var cfg Config
	if err := cfgloader.Load(ctx, &cfg, cfgloader.WithLookuper(lu)); err != nil {
		return nil, fmt.Errorf("failed to parse webhook server config: %w", err)
	}
	return &cfg, nil
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("COMMON SERVER OPTIONS")

	f.IntVar(&cli.IntVar{
		Name:   "batch-size",
		Target: &cfg.BatchSize,
		EnvVar: "BATCH_SIZE",
		Usage:  `The maximum number of records to process.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id",
		Target: &cfg.GitHubInstallID,
		EnvVar: "GITHUB_INSTALL_ID",
		Usage:  `The provisioned GitHub install reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key",
		Target: &cfg.GitHubPrivateKey,
		EnvVar: "GITHUB_PRIVATE_KEY",
		Usage:  `The GitHub app private key.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "logs-bucket-name",
		Target: &cfg.LogsBucketName,
		EnvVar: "LOGS_BUCKET_NAME",
		Usage:  `The name of the bucket to store GitHub action logs to.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-project-id",
		Target: &cfg.EventsProjectID,
		EnvVar: "EVENTS_PROJECT_ID",
		Usage:  `The project id that contains the events table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-table",
		Target: &cfg.EventsTable,
		EnvVar: "EVENTS_TABLE",
		Usage:  `The dataset.table_name of the events table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "leech-project-id",
		Target: &cfg.LeechProjectID,
		EnvVar: "LEECH_PROJECT_ID",
		Usage:  `The project id that contains the leech pipeline table.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "leech-table",
		Target: &cfg.LeechTable,
		EnvVar: "LEECH_TABLE",
		Usage:  `The dataset.table_name of the leech pipeline table.`,
	})
	return set
}
