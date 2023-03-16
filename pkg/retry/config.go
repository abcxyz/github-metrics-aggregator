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
	"strconv"

	"github.com/abcxyz/pkg/cfgloader"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the set of environment variables required
// for running the retry service.
type Config struct {
	Port string `env:"PORT,default=8081"`
}

// Validate validates the retry config after load.
func (s *Config) Validate() error {
	p, err := strconv.Atoi(s.Port)
	if err != nil {
		return fmt.Errorf("invalid port value: %w", err)
	}

	if min, max := 1, 65535; p < min || p > max {
		return fmt.Errorf("port value must be between %s and %s", strconv.Itoa(min), strconv.Itoa(max))
	}

	return nil
}

// NewConfig creates a new Config from environment variables.
func NewConfig(ctx context.Context) (*Config, error) {
	var cfg Config
	err := cfgloader.Load(ctx, &cfg, cfgloader.WithLookuper(envconfig.OsLookuper()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse retry server config: %w", err)
	}
	return &cfg, nil
}
