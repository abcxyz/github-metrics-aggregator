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
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	BatchSize        int    `env:"BATCH_SIZE,default=100"`
	EventsProjectID  string `env:"EVENTS_PROJECT_ID"`
	EventsTable      string `env:"EVENTS_TABLE"`
	GitHubAppID      string `env:"GITHUB_APP_ID"`
	GitHubInstallID  string `env:"GITHUB_INSTALL_ID"`
	GitHubPrivateKey string `env:"GITHUB_PRIVATE_KEY"`
	LeechProjectID   string `env:"LEECH_PROJECT_ID"`
	LeechTable       string `env:"LEECH_TABLE"`
	LogsBucketName   string `env:"LOGS_BUCKET_NAME"`
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
		return nil, fmt.Errorf("failed to parse pipeline config: %w", err)
	}
	return &cfg, nil
}
