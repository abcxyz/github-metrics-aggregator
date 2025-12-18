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

package retry

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v61/github"
	"github.com/sethvargo/go-gcslock"
	"google.golang.org/api/option"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/pkg/logging"
)

// eventIdentifier represents the required information used by the retry
// service for handling a GitHub event.
type eventIdentifier struct {
	eventID int64
	guid    string
}

// Datastore adheres to the interaction the retry service has with a datastore.
type Datastore interface {
	RetrieveCheckpointID(ctx context.Context, checkpointTableID, githubDomain string) (string, error)
	WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt, githubDomain string) error
	DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error)
	Close() error
}

// GitHubSource aheres to the interaction the retyr service has with GitHub APIs.
type GitHubSource interface {
	ListDeliveries(ctx context.Context, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error)
	RedeliverEvent(ctx context.Context, deliveryID int64) error
}

// RetryClientOptions encapsulate client config options as well as dependency
// implementation overrides.
type RetryClientOptions struct {
	BigQueryClientOpts      []option.ClientOption
	GCSLockClientOpts       []option.ClientOption
	DatastoreClientOverride Datastore        // used for unit testing
	GCSLockClientOverride   gcslock.Lockable // used for unit testing
	GitHubOverride          GitHubSource     // used for unit testing
	// GitHubClientCreator is used to create a new GitHub client.
	// If nil, githubclient.New is used.
	GitHubClientCreator func(context.Context, *githubclient.Config) (GitHubSource, error)
	// Now is used to get the current time.
	// If nil, time.Now is used.
	Now func() time.Time
}

// ExecuteJob runs the retry job to find and retry failed webhook events.
func ExecuteJob(ctx context.Context, cfg *Config, rco *RetryClientOptions) error {
	logger := logging.FromContext(ctx)
	nowFunc := rco.Now
	if nowFunc == nil {
		nowFunc = time.Now
	}
	now := nowFunc().UTC()

	datastore := rco.DatastoreClientOverride
	if datastore == nil {
		bq, err := NewBigQuery(ctx, cfg.BigQueryProjectID, cfg.DatasetID, rco.BigQueryClientOpts...)
		if err != nil {
			return fmt.Errorf("failed to initialize BigQuery client: %w", err)
		}
		datastore = bq
	}
	defer datastore.Close()

	gcsLock := rco.GCSLockClientOverride
	if gcsLock == nil {
		lock, err := gcslock.New(ctx, cfg.BucketName, "retry-lock", rco.GCSLockClientOpts...)
		if err != nil {
			return fmt.Errorf("failed to obtain GCS lock: %w", err)
		}
		gcsLock = lock
	}
	defer gcsLock.Close(ctx)

	githubClient := rco.GitHubOverride
	if githubClient == nil {
		var err error
		if rco.GitHubClientCreator != nil {
			githubClient, err = rco.GitHubClientCreator(ctx, &cfg.GitHub)
		} else {
			githubClient, err = githubclient.New(ctx, &cfg.GitHub)
		}
		if err != nil {
			return fmt.Errorf("failed to initialize github client: %w", err)
		}
	}

	tokenCreatedAt := nowFunc()

	if err := gcsLock.Acquire(ctx, cfg.LockTTL); err != nil {
		var lockErr *gcslock.LockHeldError
		if errors.As(err, &lockErr) {
			logger.InfoContext(ctx, "lock is already acquired by another execution", "error", lockErr.Error())
			return nil
		}
		return fmt.Errorf("failed to acquire gcs lock: %w", err)
	}

	logger.InfoContext(
		ctx,
		"looking up checkpoint",
		"checkpoint_table",
		cfg.CheckpointTableID,
		"github_domain",
		cfg.GitHubDomain,
	)
	prevCheckpoint, err := datastore.RetrieveCheckpointID(ctx, cfg.CheckpointTableID, cfg.GitHubDomain)
	if err != nil {
		return fmt.Errorf("failed to retrieve checkpoint: %w", err)
	}
	logger.InfoContext(ctx, "retrieved last checkpoint", "prev_checkpoint", prevCheckpoint)

	var totalEventCount int
	var redeliveredEventCount int
	var firstCheckpoint string
	var cursor string
	newCheckpoint := prevCheckpoint

	var failedEventsHistory []*eventIdentifier
	var found bool

	for ok := true; ok; ok = (cursor != "" && !found) {
		if rco.GitHubOverride == nil && nowFunc().Sub(tokenCreatedAt) > 4*time.Minute {
			logger.InfoContext(ctx, "refreshing github client token")
			var err error
			if rco.GitHubClientCreator != nil {
				githubClient, err = rco.GitHubClientCreator(ctx, &cfg.GitHub)
			} else {
				githubClient, err = githubclient.New(ctx, &cfg.GitHub)
			}
			if err != nil {
				return fmt.Errorf("failed to refresh github client: %w", err)
			}
			tokenCreatedAt = nowFunc()
		}

		deliveries, res, err := githubClient.ListDeliveries(ctx, &github.ListCursorOptions{
			Cursor:  cursor,
			PerPage: 100,
		})
		if err != nil {
			return fmt.Errorf("failed to list deliveries: %w", err)
		}

		if len(deliveries) == 0 {
			logger.InfoContext(ctx, "no deliveries from GitHub", "cursor", cursor)
			break
		}

		if firstCheckpoint == "" {
			firstCheckpoint = strconv.FormatInt(*deliveries[0].ID, 10)
		}

		logger.InfoContext(
			ctx,
			"retrieve deliveries from GitHub",
			"cursor",
			cursor,
			"size",
			len(deliveries),
		)
		cursor = res.Cursor

		for i := 0; i < len(deliveries); i++ {
			totalEventCount += 1
			event := deliveries[i]
			if prevCheckpoint == strconv.FormatInt(*event.ID, 10) {
				found = true
				break
			}
			if *event.StatusCode >= 200 && *event.StatusCode <= 299 {
				continue
			}
			failedEventsHistory = append(failedEventsHistory, &eventIdentifier{eventID: *event.ID, guid: *event.GUID})
		}
	}

	failedEventCount := len(failedEventsHistory)
	for i := failedEventCount - 1; failedEventCount > 0 && i >= 0; i-- {
		eventIdentifier := failedEventsHistory[i]
		if err := githubClient.RedeliverEvent(ctx, eventIdentifier.eventID); err != nil {
			var acceptedErr *github.AcceptedError
			if !errors.As(err, &acceptedErr) {
				exists, deliveryErr := datastore.DeliveryEventExists(ctx, cfg.EventsTableID, eventIdentifier.guid)
				if deliveryErr != nil {
					if newCheckpoint != prevCheckpoint {
						writeMostRecentCheckpoint(ctx, datastore, cfg.CheckpointTableID, newCheckpoint, prevCheckpoint, now, cfg.GitHubDomain)
					}
					return fmt.Errorf("failed to check if delivery event exists: %w", err)
				}
				if !exists {
					if newCheckpoint != prevCheckpoint {
						writeMostRecentCheckpoint(ctx, datastore, cfg.CheckpointTableID, newCheckpoint, prevCheckpoint, now, cfg.GitHubDomain)
					}
					return fmt.Errorf("failed to redeliver event: %w", err)
				}
			}
		}
		logger.InfoContext(ctx, "detected a failed event and successfully redelivered", "event_id", eventIdentifier.eventID)
		redeliveredEventCount += 1
		newCheckpoint = strconv.FormatInt(eventIdentifier.eventID, 10)
	}

	if firstCheckpoint == "" {
		logger.WarnContext(ctx, "ListDeliveries request did not return any deliveries, skipping checkpoint update")
	} else {
		newCheckpoint = firstCheckpoint
		logger.DebugContext(ctx, "updating checkpoint", "new_checkpoint", newCheckpoint)
		writeMostRecentCheckpoint(ctx, datastore, cfg.CheckpointTableID, newCheckpoint, prevCheckpoint, now, cfg.GitHubDomain)
	}

	logger.InfoContext(ctx, "successful",
		"total_event_count", totalEventCount,
		"failed_event_count", failedEventCount,
		"redelivered_event_count", redeliveredEventCount,
	)
	return nil
}

func writeMostRecentCheckpoint(ctx context.Context, datastore Datastore, checkpointTableID, newCheckpoint, prevCheckpoint string, now time.Time, githubDomain string) {
	logging.FromContext(ctx).InfoContext(ctx, "write new checkpoint",
		"prev_checkpoint", prevCheckpoint,
		"new_checkpoint", newCheckpoint)
	createdAt := now.Format(time.DateTime)
	if err := datastore.WriteCheckpointID(ctx, checkpointTableID, newCheckpoint, createdAt, githubDomain); err != nil {
		logging.FromContext(ctx).ErrorContext(ctx, "failed to write checkpoint", "error", err)
	}
}
