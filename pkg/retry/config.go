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
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set of environment variables required
// for running the retry service.
type Config struct {
	AppID             string        `env:"GITHUB_APP_ID,required"`
	BigQueryProjectID string        `env:"BIG_QUERY_PROJECT_ID,default=$PROJECT_ID"`
	BucketURL         string        `env:"BUCKET_URL,required"`
	CheckpointTableID string        `env:"CHECKPOINT_TABLE_ID,required"`
	DatasetID         string        `env:"DATASET_ID,required"`
	LockTTLClockSkew  time.Duration `env:"LOCK_TTL_CLOCK_SKEW,default=10s"`
	LockTTL           time.Duration `env:"LOCK_TTL,default=5m"`
	ProjectID         string        `env:"PROJECT_ID,required"`
	Port              string        `env:"PORT,default=8080"`
	WebhookID         string        `env:"GITHUB_WEBHOOK_ID,required"`
}

// Validate validates the retry config after load.
func (cfg *Config) Validate() error {
	if cfg.AppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
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

	if cfg.WebhookID == "" {
		return fmt.Errorf("GITHUB_WEBHOOK_ID is required")
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
