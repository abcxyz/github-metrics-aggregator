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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-github/v52/github"
	"github.com/sethvargo/go-gcslock"
)

const (
	acceptedMessage        = "Accepted"
	errAcquireLock         = "Failed to acquire GCS lock."
	errDeliveryEventExists = "Failed to check if event exists"
	errWriteCheckpoint     = "Failed to write checkpoint."
	errRetrieveCheckpoint  = "Failed to retrieve checkpoint."
	errCallingGitHub       = "Failed to call GitHub."
)

// eventIdentifier represents the required information used by the retry
// service for handling a GitHub event.
type eventIdentifier struct {
	eventID int64
	guid    string
}

// handleRetry handles calling GitHub APIs to search and retry for failed
// events.
func (s *Server) handleRetry() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		ctx := r.Context()
		logger := logging.FromContext(ctx)

		if err := s.gcsLock.Acquire(ctx, s.lockTTL); err != nil {
			var lockErr *gcslock.LockHeldError
			if errors.As(err, &lockErr) {
				logger.InfoContext(ctx, "lock is already acquired by another execution",
					"code", http.StatusOK,
					"body", errAcquireLock,
					"method", "Acquire",
					"error", lockErr.Error(),
				)

				// unable to obtain the lock, return a 200 so the scheduler doesnt attempt to reinvoke
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, http.StatusText(http.StatusOK))
				return
			}

			logger.ErrorContext(ctx, "failed to call cloud storage",
				"code", http.StatusInternalServerError,
				"body", errAcquireLock,
				"method", "Acquire",
				"error", err.Error())

			// unknown error
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// read the last checkpoint from checkpoint table
		prevCheckpoint, err := s.datastore.RetrieveCheckpointID(ctx, s.checkpointTableID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to call RetrieveCheckpointID",
				"code", http.StatusInternalServerError,
				"body", errRetrieveCheckpoint,
				"method", "RetrieveCheckpointID",
				"error", err,
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.ErrorContext(ctx, "retrieved last checkpoint", "prevCheckpoint", prevCheckpoint)

		var totalEventCount int
		var redeliveredEventCount int
		var firstCheckpoint string
		var cursor string
		newCheckpoint := prevCheckpoint

		// store all observed failures in memory from the latest event up to the prevCheckpoint
		var failedEventsHistory []*eventIdentifier
		var found bool

		// the first run of this service will not have a cursor therefore we must
		// ensure we run the loop at least once
		for ok := true; ok; ok = (cursor != "" && !found) {
			// call list deliveries API, first call is intentionally an empty string
			deliveries, res, err := s.github.ListDeliveries(ctx, &github.ListCursorOptions{
				Cursor:  cursor,
				PerPage: 100,
			})
			if err != nil {
				logger.ErrorContext(ctx, "failed to call ListDeliveries",
					"code", http.StatusInternalServerError,
					"body", errCallingGitHub,
					"method", "RedeliverEvent",
					"error", err,
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			// in anticipation of the happy path, store the first event to advance the
			// cursor
			if firstCheckpoint == "" {
				firstCheckpoint = strconv.FormatInt(*deliveries[0].ID, 10)
			}

			logger.ErrorContext(ctx, "retrieve deliveries from GitHub", "cursor", cursor, "size", len(deliveries))

			// update the cursor
			cursor = res.Cursor

			// for each failed delivery, redeliver
			for i := 0; i < len(deliveries); i++ {
				// append to the total events counter
				totalEventCount += 1

				event := deliveries[i]

				// reached the last checkpoint, all events equal to and older than this
				// one have already been processed
				if prevCheckpoint == strconv.FormatInt(*event.ID, 10) {
					found = true
					break
				}

				// check payload and see if its been successfully delivered, if so skip over it
				if *event.StatusCode >= 200 && *event.StatusCode <= 299 {
					continue
				}

				failedEventsHistory = append(failedEventsHistory, &eventIdentifier{eventID: *event.ID, guid: *event.GUID})
			}
		}

		failedEventCount := len(failedEventsHistory)

		// work backwards from the list of failed events then attempt redelivery and
		// increment the newCheckpoint in an effort to close the gap to the most
		// recent event, this should alleviate pressure on future runs
		for i := failedEventCount - 1; failedEventCount > 0 && i >= 0; i-- {
			eventIdentifier := failedEventsHistory[i]

			if err := s.github.RedeliverEvent(ctx, eventIdentifier.eventID); err != nil {
				var acceptedErr *github.AcceptedError
				if !errors.As(err, &acceptedErr) {
					// found an unaccepted error, check if its already in the events table
					exists, err := s.datastore.DeliveryEventExists(ctx, s.eventsTableID, eventIdentifier.guid)
					if err != nil {
						logger.ErrorContext(ctx, "failed to call BigQuery",
							"method", "DeliveryEventExists",
							"code", http.StatusInternalServerError,
							"body", errDeliveryEventExists,
							"error", err,
						)

						if newCheckpoint != prevCheckpoint {
							s.writeMostRecentCheckpoint(ctx, w, newCheckpoint, prevCheckpoint, now,
								totalEventCount, failedEventCount, redeliveredEventCount)
						}

						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					if !exists {
						logger.ErrorContext(ctx, "failed to redeliver event, stop processing",
							"code", http.StatusInternalServerError,
							"body", errCallingGitHub,
							"method", "RedeliverEvent",
							"guid", eventIdentifier.guid,
							"error", err,
							"totalEventCount", totalEventCount,
							"failedEventCount", failedEventCount,
						)

						if newCheckpoint != prevCheckpoint {
							s.writeMostRecentCheckpoint(ctx, w, newCheckpoint, prevCheckpoint, now,
								totalEventCount, failedEventCount, redeliveredEventCount)
						}

						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
				}
			}

			logger.InfoContext(ctx, "detected a failed event and successfully redelivered", "eventID", eventIdentifier.eventID)
			redeliveredEventCount += 1

			newCheckpoint = strconv.FormatInt(eventIdentifier.eventID, 10)
		}

		// advance the checkpoint to the first entry read on this run to avoid
		// redundant processing
		newCheckpoint = firstCheckpoint

		s.writeMostRecentCheckpoint(ctx, w, newCheckpoint, prevCheckpoint, now,
			totalEventCount, failedEventCount, redeliveredEventCount)

		logger.InfoContext(ctx, "successful",
			"code", http.StatusAccepted,
			"body", acceptedMessage,
			"totalEventCount", totalEventCount,
			"failedEventCount", failedEventCount,
			"redeliveredEventCount", redeliveredEventCount,
		)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, http.StatusText(http.StatusAccepted))
	})
}

// writeMostRecentCheckpoint is a helper function to write to the checkpoint
// table with the last successfully processed checkpoint denoted by
// newCheckpoint.
func (s *Server) writeMostRecentCheckpoint(ctx context.Context, w http.ResponseWriter,
	newCheckpoint, prevCheckpoint string, now time.Time, totalEventCount, failedEventCount, redeliveredEventCount int,
) {
	logging.FromContext(ctx).InfoContext(ctx, "write new checkpoint", "prevCheckpoint", prevCheckpoint, "newCheckpoint", newCheckpoint)
	createdAt := now.Format(time.DateTime)
	if err := s.datastore.WriteCheckpointID(ctx, s.checkpointTableID, newCheckpoint, createdAt); err != nil {
		logging.FromContext(ctx).ErrorContext(ctx, "failed to call WriteCheckpointID",
			"code", http.StatusInternalServerError,
			"body", errWriteCheckpoint,
			"method", "RedeliverEvent",
			"error", err,
			"totalEventCount", totalEventCount,
			"failedEventCount", failedEventCount,
			"redeliveredEventCount", redeliveredEventCount,
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
