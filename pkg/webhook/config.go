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
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	BigQueryID           string `env:"BIG_QUERY_ID,required"`
	DatasetID            string `env:"DATASET_ID,required"`
	EventsTableID        string `env:"EVENTS_TABLE_ID,required"`
	FailureEventsTableID string `env:"FAILURE_EVENTS_TABLE_ID,required"`
	Port                 string `env:"PORT,default=8080"`
	ProjectID            string `env:"PROJECT_ID,required"`
	RetryLimit           int    `env:"RETRY_LIMIT,required"`
	TopicID              string `env:"TOPIC_ID,required"`
	WebhookSecret        string `env:"WEBHOOK_SECRET,required"`
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
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

	if cfg.TopicID == "" {
		return fmt.Errorf("TOPIC_ID is required")
	}

	if cfg.WebhookSecret == "" {
		return fmt.Errorf("WEBHOOK_SECRET is required")
	}

	if cfg.RetryLimit <= 0 {
		return fmt.Errorf("RETRY_LIMIT is required and must be greater than 0")
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
