// Copyright 2024 The Authors (see AUTHORS file)
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

package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/pkg/cli"
)

// Config defines the set of environment variables required
// for running the artifact job.
type Config struct {
	GitHub githubclient.Config

	// ProjectID is the project id where the tables live.
	ProjectID string

	// DatasetID is the dataset id where the tables live.
	DatasetID string

	// PushEventsTableID is the table_name of the push events table.
	PushEventsTableID string

	// CommitReviewStatusTableID is the table_name of the commit_review_status table.
	CommitReviewStatusTableID string

	// IssuesTableID is the table_name of the issues table.
	IssuesTableID string
}

// Validate validates the artifacts config after load.
func (cfg *Config) Validate(ctx context.Context) error {
	var merr error

	merr = errors.Join(cfg.GitHub.Validate(ctx))

	if cfg.PushEventsTableID == "" {
		merr = errors.Join(merr, fmt.Errorf("PUSH_EVENTS_TABLE_ID is required"))
	}

	if (cfg.CommitReviewStatusTableID) == "" {
		merr = errors.Join(merr, fmt.Errorf("COMMIT_REVIEW_STATUS_TABLE_ID is required"))
	}

	if (cfg.IssuesTableID) == "" {
		merr = errors.Join(merr, fmt.Errorf("ISSUES_TABLE_ID is required"))
	}

	if cfg.ProjectID == "" {
		merr = errors.Join(merr, fmt.Errorf("PROJECT_ID is required"))
	}

	if cfg.DatasetID == "" {
		merr = errors.Join(merr, fmt.Errorf("DATASET_ID is required"))
	}

	return merr
}

// ToFlags binds the config to the [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	cfg.GitHub.ToFlags(set)

	f := set.NewSection("COMMON OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "push-events-table-id",
		Target:  &cfg.PushEventsTableID,
		EnvVar:  "PUSH_EVENTS_TABLE_ID",
		Usage:   `The push_events table ID within the dataset.`,
		Example: "retry-lock-xxxx",
	})

	f.StringVar(&cli.StringVar{
		Name:   "commit-review-status-table-id",
		Target: &cfg.CommitReviewStatusTableID,
		EnvVar: "COMMIT_REVIEW_STATUS_TABLE_ID",
		Usage:  `The commit_review_status table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "issues-table-id",
		Target: &cfg.IssuesTableID,
		EnvVar: "ISSUES_TABLE_ID",
		Usage:  `The issues table ID within the dataset.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "dataset-id",
		Target: &cfg.DatasetID,
		EnvVar: "DATASET_ID",
		Usage:  `BigQuery dataset ID.`,
	})

	return set
}
