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

package review

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the artifact job.
type Config struct {
	GitHubAppID            string `env:"GITHUB_APP_ID,required"`             // The GitHub App ID
	GitHubInstallID        string `env:"GITHUB_INSTALL_ID,required"`         // The provisioned GitHub App Installation reference
	GitHubPrivateKeySecret string `env:"GITHUB_PRIVATE_KEY_SECRET,required"` // The secret name & version containing the GitHub App private key

	ProjectID string `env:"PROJECT_ID,required"` // The project id where the tables live
	DatasetID string `env:"DATASET_ID,required"` // The dataset id where the tables live

	PushEventsTableID         string `env:"PUSH_EVENTS_TABLE_ID,required"`          // The table_name of the push events table
	CommitReviewStatusTableID string `env:"COMMIT_REVIEW_STATUS_TABLE_ID,required"` // The table_name of the commit_review_status table
	IssuesTableID             string `env:"ISSUES_TABLE_ID,required"`               // The table_name of the issues table
}

// Validate validates the artifacts config after load.
func (cfg *Config) Validate() error {
	if cfg.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}
	if cfg.GitHubInstallID == "" {
		return fmt.Errorf("GITHUB_INSTALL_ID is required")
	}

	if cfg.GitHubPrivateKeySecret == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY_SECRET is required")
	}

	if cfg.PushEventsTableID == "" {
		return fmt.Errorf("PUSH_EVENTS_TABLE_ID is required")
	}

	if (cfg.CommitReviewStatusTableID) == "" {
		return fmt.Errorf("COMMIT_REVIEW_STATUS_TABLE_ID is required")
	}

	if (cfg.IssuesTableID) == "" {
		return fmt.Errorf("ISSUES_TABLE_ID is required")
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
		Name:   "github-app-id",
		Target: &cfg.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-install-id",
		Target: &cfg.GitHubInstallID,
		EnvVar: "GITHUB_INSTALL_ID",
		Usage:  `The provisioned GitHub App installation ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key-secret",
		Target: &cfg.GitHubPrivateKeySecret,
		EnvVar: "GITHUB_PRIVATE_KEY_SECRET",
		Usage:  `The secret name & version containing the GitHub App private key.`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "push-events-table-id",
		Target:  &cfg.PushEventsTableID,
		EnvVar:  "PUSH_EVENTS_TABLE_ID",
		Usage:   `The push_events table ID within the dataset.`,
		Example: "retry-lock-xxxx",
	})

	f.StringVar(&cli.StringVar{
		Name:   "commit-review-status-table-id",
		Target: &cfg.CommitReviewStatusTableID,
		EnvVar: "COMMIT_REVIEW_STATUS_TABLE_ID",
		Usage:  `The commit_review_status table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "issues-table-id",
		Target: &cfg.IssuesTableID,
		EnvVar: "ISSUES_TABLE_ID",
		Usage:  `The issues table ID within the dataset.`,
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

	return set
}
