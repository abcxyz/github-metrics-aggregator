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

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set of environment variables required
// for running the retry service.
type Config struct {
	AppID              string `env:"GITHUB_APP_ID,required"`
	BigQueryID         string `env:"BIG_QUERY_ID,required"`
	BucketURL          string `env:"BUCKET_URL,required"`
	LockTTLClockSkewMS int    `env:"LOCK_TTL_CLOCK_SKEW_MS,required"`
	LockTTLMinutes     int    `env:"LOCK_TTL_MINUTES,required"`
	ProjectID          string `env:"PROJECT_ID,required"`
	Port               string `env:"PORT,default=8081"`
	WebhookID          string `env:"GITHUB_WEBHOOK_ID,required"`
}

// Validate validates the retry config after load.
func (cfg *Config) Validate() error {
	if len(cfg.AppID) == 0 {
		return fmt.Errorf("GITHUB_APP_ID is empty and requires a value")
	}

	if len(cfg.BigQueryID) == 0 {
		return fmt.Errorf("BIG_QUERY_ID is empty and requires a value")
	}

	if len(cfg.BucketURL) == 0 {
		return fmt.Errorf("BUCKET_URL is empty and requires a value")
	}

	if cfg.LockTTLClockSkewMS < 0 {
		return fmt.Errorf("LockTTLClockSkewMS must be a positive value, got: %v", cfg.LockTTLClockSkewMS)
	}

	if cfg.LockTTLMinutes < 0 {
		return fmt.Errorf("LockTTLMinutes must be a positive value, got: %v", cfg.LockTTLMinutes)
	}

	if len(cfg.ProjectID) == 0 {
		return fmt.Errorf("PROJECT_ID is empty and requires a value")
	}

	if len(cfg.WebhookID) == 0 {
		return fmt.Errorf("GITHUB_WEBHOOK_ID is empty and requires a value")
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
