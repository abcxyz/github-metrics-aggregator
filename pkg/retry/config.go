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

package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set of environment variables required
// for running the retry service.
type Config struct {
	GitHubAppID       string        `env:"GITHUB_APP_ID,required"`
	GitHubInstallID   string        `env:"GITHUB_INSTALL_ID,required"`
	BigQueryProjectID string        `env:"BIG_QUERY_PROJECT_ID,default=$PROJECT_ID"`
	BucketURL         string        `env:"BUCKET_URL,required"`
	CheckpointTableID string        `env:"CHECKPOINT_TABLE_ID,required"`
	DatasetID         string        `env:"DATASET_ID,required"`
	LockTTLClockSkew  time.Duration `env:"LOCK_TTL_CLOCK_SKEW,default=10s"`
	LockTTL           time.Duration `env:"LOCK_TTL,default=5m"`
	ProjectID         string        `env:"PROJECT_ID,required"`
	Port              string        `env:"PORT,default=8080"`
}

// Validate validates the retry config after load.
func (cfg *Config) Validate() error {
	if cfg.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}

	if cfg.GitHubInstallID == "" {
		return fmt.Errorf("GITHUB_INSTALL_ID is required")
	}

	if cfg.BucketURL == "" {
		return fmt.Errorf("BUCKET_URL is required")
	}

	if (cfg.CheckpointTableID) == "" {
		return fmt.Errorf("CHECKPOINT_TABLE_ID is required")
	}

	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID is required")
	}

	if cfg.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
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

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("COMMON SERVER OPTIONS")

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
		Name:   "big-query-project-id",
		Target: &cfg.BigQueryProjectID,
		EnvVar: "BIG_QUERY_PROJECT_ID",
		Usage:  `The project ID where your BigQuery instance exists in.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "bucket-url",
		Target: &cfg.BucketURL,
		EnvVar: "BUCKET_URL",
		Usage:  `The URL for the bucket that holds the lock to enforce synchronous processing of the retry service.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "checkpoint-table-id",
		Target: &cfg.CheckpointTableID,
		EnvVar: "CHECKPOINT_TABLE_ID",
		Usage:  `The checkpoint table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "dataset-id",
		Target: &cfg.DatasetID,
		EnvVar: "DATASET_ID",
		Usage:  `The dataset ID within the BigQuery instance.`,
	})

	f.DurationVar(&cli.DurationVar{
		Name:    "lock-ttl-clock-skew",
		Target:  &cfg.LockTTLClockSkew,
		EnvVar:  "LOCK_TTL_CLOCK_SKEW",
		Default: 10 * time.Second,
		Usage:   "Duration to account for clock drift.",
	})

	f.DurationVar(&cli.DurationVar{
		Name:    "lock-ttl",
		Target:  &cfg.LockTTL,
		EnvVar:  "LOCK_TTL",
		Default: 5 * time.Minute,
		Usage:   "Duration for a lock to be active until it is allowed to be taken.",
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "port",
		Target:  &cfg.Port,
		EnvVar:  "PORT",
		Default: "8080",
		Usage:   `The port the retry server listens to.`,
	})

	return set
}
