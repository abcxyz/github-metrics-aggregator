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

package webhook

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
	BigQueryProjectID    string `env:"BIG_QUERY_PROJECT_ID,default=$PROJECT_ID"`
	DatasetID            string `env:"DATASET_ID,required"`
	EventsTableID        string `env:"EVENTS_TABLE_ID,required"`
	FailureEventsTableID string `env:"FAILURE_EVENTS_TABLE_ID,required"`
	Port                 string `env:"PORT,default=8080"`
	ProjectID            string `env:"PROJECT_ID,required"`
	RetryLimit           int    `env:"RETRY_LIMIT,required"`
	EventsTopicID        string `env:"EVENTS_TOPIC_ID,required"`
	DLQEventsTopicID     string `env:"DLQ_EVENTS_TOPIC_ID,required"`
	GitHubWebhookSecret  string `env:"GITHUB_WEBHOOK_SECRET,required"`
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID is required")
	}

	if cfg.EventsTableID == "" {
		return fmt.Errorf("EVENTS_TABLE_ID is required")
	}

	if cfg.FailureEventsTableID == "" {
		return fmt.Errorf("FAILURE_EVENTS_TABLE_ID is required")
	}

	// TODO: get project from compute metadata server if required in future
	if cfg.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
	}

	if cfg.RetryLimit <= 0 {
		return fmt.Errorf("RETRY_LIMIT is required and must be greater than 0")
	}

	if cfg.EventsTopicID == "" {
		return fmt.Errorf("EVENTS_TOPIC_ID is required")
	}

	if cfg.DLQEventsTopicID == "" {
		return fmt.Errorf("DLQ_EVENTS_TOPIC_ID is required")
	}

	if cfg.GitHubWebhookSecret == "" {
		return fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
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

	f.StringVar(&cli.StringVar{
		Name:   "big-query-project-id",
		Target: &cfg.BigQueryProjectID,
		EnvVar: "BIG_QUERY_PROJECT_ID",
		Usage:  `The project ID where your BigQuery instance exists in.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "dataset-id",
		Target: &cfg.DatasetID,
		EnvVar: "DATASET_ID",
		Usage:  `The dataset ID within the BigQuery instance.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-table-id",
		Target: &cfg.EventsTableID,
		EnvVar: "EVENTS_TABLE_ID",
		Usage:  `The events table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "failure-events-table-id",
		Target: &cfg.FailureEventsTableID,
		EnvVar: "FAILURE_EVENTS_TABLE_ID",
		Usage:  `The failure events table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "port",
		Target:  &cfg.Port,
		EnvVar:  "PORT",
		Default: "8080",
		Usage:   `The port the retry server listens to.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID.`,
	})

	f.IntVar(&cli.IntVar{
		Name:   "retry-limit",
		Target: &cfg.RetryLimit,
		EnvVar: "RETRY_LIMIT",
		Usage:  `The maximum retry limit before giving up on an event.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "events-topic-id",
		Target: &cfg.EventsTopicID,
		EnvVar: "EVENTS_TOPIC_ID",
		Usage:  `Google PubSub topic ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "dlq-events-topic-id",
		Target: &cfg.DLQEventsTopicID,
		EnvVar: "DLQ_EVENTS_TOPIC_ID",
		Usage:  `Google PubSub topic ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-webhook-secret",
		Target: &cfg.GitHubWebhookSecret,
		EnvVar: "GITHUB_WEBHOOK_SECRET",
		Usage:  `GitHub webhook secret.`,
	})

	return set
}
