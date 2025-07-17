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
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the retry service.
type Config struct {
	GitHubEnterpriseServerURL string        `env:"GITHUB_ENTERPRISE_SERVER_URL"`
	GitHubAppID               string        `env:"GITHUB_APP_ID,required"`
	GitHubPrivateKey          string        `env:"GITHUB_PRIVATE_KEY,required"`
	BigQueryProjectID         string        `env:"BIG_QUERY_PROJECT_ID,default=$PROJECT_ID"`
	BucketName                string        `env:"BUCKET_NAME,required"`
	CheckpointTableID         string        `env:"CHECKPOINT_TABLE_ID,required"`
	EventsTableID             string        `env:"EVENTS_TABLE_ID,required"`
	DatasetID                 string        `env:"DATASET_ID,required"`
	LockTTLClockSkew          time.Duration `env:"LOCK_TTL_CLOCK_SKEW,default=10s"`
	LockTTL                   time.Duration `env:"LOCK_TTL,default=5m"`
	ProjectID                 string        `env:"PROJECT_ID,required"`
	Port                      string        `env:"PORT,default=8080"`
}

// Validate validates the retry config after load.
func (cfg *Config) Validate() error {
	if cfg.GitHubEnterpriseServerURL != "" && !strings.HasPrefix(cfg.GitHubEnterpriseServerURL, "https://") {
		return fmt.Errorf("GITHUB_ENTERPRISE_SERVER_URL does not start with \"https://\"")
	}

	if cfg.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}

	if cfg.GitHubPrivateKey == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY is required")
	}

	if cfg.BucketName == "" {
		return fmt.Errorf("BUCKET_NAME is required")
	}

	if (cfg.CheckpointTableID) == "" {
		return fmt.Errorf("CHECKPOINT_TABLE_ID is required")
	}

	if (cfg.EventsTableID) == "" {
		return fmt.Errorf("EVENTS_TABLE_ID is required")
	}

	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID is required")
	}

	if cfg.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
	}

	// Given this Validate function runs after the ToFlags function, this fallback
	// is done in case the user has not provided a BIG_QUERY_PROJECT_ID.
	if cfg.BigQueryProjectID == "" {
		cfg.BigQueryProjectID = cfg.ProjectID
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
		Name:   "github-enterprise-server_url",
		Target: &cfg.GitHubEnterpriseServerURL,
		EnvVar: "GITHUB_ENTERPRISE_SERVER_URL",
		Usage:  `The GitHub Enterprise Server instance URL, format "https://[hostname]"`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App reference.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key",
		Target: &cfg.GitHubPrivateKey,
		EnvVar: "GITHUB_PRIVATE_KEY",
		Usage:  `The private key generated to call GitHub.`,
	})

	// This will default to projectID in the Validate function
	// and is intentionally not done here.
	f.StringVar(&cli.StringVar{
		Name:   "big-query-project-id",
		Target: &cfg.BigQueryProjectID,
		EnvVar: "BIG_QUERY_PROJECT_ID",
		Usage:  `The project ID where your BigQuery instance exists in.`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "bucket-name",
		Target:  &cfg.BucketName,
		EnvVar:  "BUCKET_NAME",
		Usage:   `The name of the bucket that holds the lock to enforce synchronous processing of the retry service.`,
		Example: "retry-lock-xxxx",
	})

	f.StringVar(&cli.StringVar{
		Name:   "checkpoint-table-id",
		Target: &cfg.CheckpointTableID,
		EnvVar: "CHECKPOINT_TABLE_ID",
		Usage:  `The checkpoint table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-table-id",
		Target: &cfg.EventsTableID,
		EnvVar: "EVENTS_TABLE_ID",
		Usage:  `The events table ID within the dataset.`,
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
