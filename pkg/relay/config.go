// Copyright 2025 The Authors (see AUTHORS file)
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

package relay

import (
	"errors"
	"fmt"
	"time"

	"github.com/abcxyz/pkg/cli"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	Port           string
	ProjectID      string
	RelayTopicID   string
	RelayProjectID string
	PubSubTimeout  time.Duration
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
	var merr error

	if cfg.ProjectID == "" {
		merr = errors.Join(merr, fmt.Errorf("PROJECT_ID is required"))
	}

	if cfg.PubSubTimeout <= 0 {
		merr = errors.Join(merr, fmt.Errorf("PUBSUB_TIMEOUT must be positive"))
	}

	return merr
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("RELAY OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "port",
		Target:  &cfg.Port,
		EnvVar:  "PORT",
		Default: "8080",
		Usage:   `The port the relay server listens to.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID where this service runs.`,
	})
	f.StringVar(&cli.StringVar{
		Name:   "relay-topic-id",
		Target: &cfg.RelayTopicID,
		EnvVar: "RELAY_TOPIC_ID",
		Usage:  `Google PubSub topic ID.`,
	})
	f.StringVar(&cli.StringVar{
		Name:   "relay-project-id",
		Target: &cfg.RelayProjectID,
		EnvVar: "RELAY_PROJECT_ID",
		Usage:  `Google Cloud project ID where the relay topic lives.`,
	})

	f.DurationVar(&cli.DurationVar{
		Name:    "pubsub-timeout",
		Target:  &cfg.PubSubTimeout,
		EnvVar:  "PUBSUB_TIMEOUT",
		Default: 10 * time.Second,
		Usage:   `The timeout for PubSub requests.`,
	})

	return set
}
