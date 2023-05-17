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
				logger.Infow("lock is already acquired by another execution",
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

			logger.Errorf("failed to call cloud storage",
				"code", http.StatusInternalServerError,
				"body", errAcquireLock,
				"method", "Acquire",
				"error", err.Error())

			// unknown error
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// read the last checkpoint from checkpoint table
		lastCheckpoint, err := s.datastore.RetrieveCheckpointID(ctx, s.checkpointTableID)
		if err != nil {
			logger.Errorw("failed to call RetrieveCheckpointID",
				"code", http.StatusInternalServerError,
				"body", errRetrieveCheckpoint,
				"method", "RetrieveCheckpointID",
				"error", err,
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Infow("retrieved last checkpoint", "lastCheckpoint", lastCheckpoint)

		// for each run track the total number of events retrieve and failure events
		var totalEventCount int
		var failedEventCount int
		var newCheckpoint *int64
		var cursor string
		var processed bool

		// the first run of this service will not have a cursor therefore we must
		// ensure we run the loop at least once
		for ok := true; ok; ok = (cursor != "") {
			// call list deliveries API, first call is intentionally an empty string
			deliveries, res, err := s.github.ListDeliveries(ctx, &github.ListCursorOptions{
				Cursor:  cursor,
				PerPage: 100,
			})
			if err != nil {
				logger.Errorw("failed to call ListDeliveries",
					"code", http.StatusInternalServerError,
					"body", errCallingGitHub,
					"method", "RedeliverEvent",
					"error", err,
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			logger.Infow("retrieve deliveries from GitHub", "cursor", cursor, "size", len(deliveries))

			// update the cursor
			cursor = res.Cursor

			// for each failed delivery, redeliver
			for i := 0; i < len(deliveries); i++ {
				// append to the total events counter
				totalEventCount += 1

				event := deliveries[i]

				// reached the last successful checkpoint, all events equal to and older
				// than this one have already been processed
				if lastCheckpoint == strconv.FormatInt(*event.ID, 10) {
					processed = true
					break
				}

				// check payload and see if its been successfully delivered, if so skip over it
				if *event.StatusCode >= 200 && *event.StatusCode <= 299 {
					continue
				}

				// append to the total failure events counter
				failedEventCount += 1

				logger.Infow("redeliver failed event", "eventID", *event.ID, "guid", *event.GUID)

				if err := s.github.RedeliverEvent(ctx, *event.ID); err != nil {
					var acceptedErr *github.AcceptedError
					if errors.As(err, &acceptedErr) {
						logger.Infow("skipping redeliver event because it has already been submitted to GitHub",
							"eventID", *event.ID,
							"guid", *event.GUID,
							"error", err,
						)
						continue
					}

					// found an unaccepted error, check if its already in the events table
					exists, err := s.datastore.DeliveryEventExists(ctx, s.eventsTableID, *event.GUID)
					if err != nil {
						logger.Errorw("failed to call BigQuery",
							"method", "DeliveryEventExists",
							"code", http.StatusInternalServerError,
							"body", errDeliveryEventExists,
							"error", err,
						)
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}

					if exists {
						logger.Infow("skipping redeliver event because it has already been processed", "eventID", *event.ID, "guid", *event.GUID)
						continue
					}

					logger.Errorw("failed to call RedeliverEvent, stop processing",
						"code", http.StatusInternalServerError,
						"body", errCallingGitHub,
						"method", "RedeliverEvent",
						"event", *event,
						"guid", *event.GUID,
						"error", err,
						"totalEventCount", totalEventCount,
						"failedEventCount", failedEventCount,
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}

			// every event from newCheckpoint to lastCheckpoint has been processed,
			// overwrite the checkpoint
			if processed {
				logger.Infow("write new checkpoint", "lastCheckpoint", lastCheckpoint, "newCheckpoint", newCheckpoint)
				deliveryID := strconv.FormatInt(*newCheckpoint, 10)
				createdAt := now.Format(time.DateTime)
				if err := s.datastore.WriteCheckpointID(ctx, s.checkpointTableID, deliveryID, createdAt); err != nil {
					logger.Errorw("failed to call WriteCheckpointID",
						"code", http.StatusInternalServerError,
						"body", errWriteCheckpoint,
						"method", "RedeliverEvent",
						"error", err,
						"totalEventCount", totalEventCount,
						"failedEventCount", failedEventCount,
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}
		}

		logger.Infow("successful",
			"code", http.StatusAccepted,
			"body", acceptedMessage,
			"totalEventCount", totalEventCount,
			"failedEventCount", failedEventCount,
		)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, http.StatusText(http.StatusAccepted))
	})
}
