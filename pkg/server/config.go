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

package server

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"github.com/abcxyz/pkg/cfgloader"
	"github.com/sethvargo/go-envconfig"
)

// ServiceConfig defines the set over environment variables required
// for running this application.
type ServiceConfig struct {
	Port          string `env:"PORT,default=8080"`
	ProjectID     string `env:"PROJECT_ID,default=$PROJECT_ID"`
	TopicID       string `env:"TOPIC_ID,required"`
	WebhookSecret string `env:"WEBHOOK_SECRET,required"`
}

// Validate validates the service config after load.
func (s *ServiceConfig) Validate() error {
	if len(s.ProjectID) == 0 {
		projectID, err := resolveDefaultProjectFunc()
		if err != nil {
			return fmt.Errorf("failed to load default project ID: %w", err)
		}
		s.ProjectID = projectID
	}

	if len(s.WebhookSecret) == 0 {
		return fmt.Errorf("WEBHOOK_SECRET is empty and requires a value")
	}

	return nil
}

// NewConfig creates a new ServiceConfig from environment variables.
func NewConfig(ctx context.Context) (*ServiceConfig, error) {
	var cfg ServiceConfig
	err := cfgloader.Load(ctx, &cfg, cfgloader.WithLookuper(envconfig.OsLookuper()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}
	return &cfg, nil
}

// resolveDefaultProjectFunc gets the Google Cloud project id from the compute metadata server.
func resolveDefaultProjectFunc() (string, error) {
	c := metadata.NewClient(nil)
	project, err := c.ProjectID()
	if err != nil {
		return "", fmt.Errorf("failed to get default project id: %w", err)
	}

	return project, nil
}
