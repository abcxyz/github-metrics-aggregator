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

// Package cli implements the commands for the github-metrics-aggregator CLI.
package cli

import (
	"context"

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/cli"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name:    "github-metrics-aggregator",
		Version: version.HumanVersion,
		Commands: map[string]cli.CommandFactory{
			"retry": func() cli.Command {
				return &cli.RootCommand{
					Name:        "retry",
					Description: "Perform retry operations",
					Commands: map[string]cli.CommandFactory{
						"server": func() cli.Command {
							return &RetryServerCommand{}
						},
					},
				}
			},
			"webhook": func() cli.Command {
				return &cli.RootCommand{
					Name:        "webhook",
					Description: "Perform webhook operations",
					Commands: map[string]cli.CommandFactory{
						"server": func() cli.Command {
							return &WebhookServerCommand{}
						},
					},
				}
			},
			"artifact": func() cli.Command {
				return &cli.RootCommand{
					Name:        "artifact",
					Description: "Execute the artifact ingestion job",
					Commands: map[string]cli.CommandFactory{
						"ingest": func() cli.Command {
							return &ArtifactJobCommand{}
						},
					},
				}
			},
		},
	}
}

// Run executes the CLI.
func Run(ctx context.Context, args []string) error {
	return rootCmd().Run(ctx, args) //nolint:wrapcheck // Want passthrough
}
