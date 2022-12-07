// Copyright 2022 GitHub Metrics Aggregator authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"github.com/sethvargo/go-envconfig"
)

var (
	projectIDEnvKey = "PROJECT_ID"
)

// Config is the server config.
type Config struct {
	Port          string `env:"PORT,default=8080"`
	ProjectID     string `env:"PROJECT_ID,default=$PROJECT_ID"`
	TopicID       string `env:"TOPIC_ID,required"`
	WebhookSecret string `env:"WEBHOOK_SECRET,required"`
}

// NewConfig handles creating a new server config and registering custom env mutator functions
func NewConfig(ctx context.Context) (*Config, error) {
	var config Config
	err := envconfig.ProcessWith(ctx, &config, envconfig.OsLookuper(), resolveDefaultProjectFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	if len(config.WebhookSecret) == 0 {
		return nil, fmt.Errorf("WEBHOOK_SECRET is empty and requires a value")
	}

	return &config, nil
}

// resolveDefaultProjectFunc gets the Google Cloud project id from the compute metadata server
// if project id is not provided
func resolveDefaultProjectFunc(ctx context.Context, key, value string) (string, error) {
	if key == projectIDEnvKey && len(value) == 0 {
		project := ""

		c := metadata.NewClient(nil)
		project, err := c.ProjectID()
		if err != nil {
			return "", fmt.Errorf("failed to get default project id: %w", err)
		}

		return project, nil
	}

	return value, nil
}
