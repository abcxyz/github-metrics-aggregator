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
	"time"

	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-github/v52/github"
	"github.com/sethvargo/go-gcslock"
)

const (
	acceptedMessage       = "Accepted"
	errAcquireLock        = "Failed to acquire GCS lock."
	errWriteCheckpoint    = "Failed to write checkpoint."
	errRetrieveCheckpoint = "Failed to retrieve checkpoint."
	errCallingGitHub      = "Failed to call GitHub."
)

// handleRetry handles calling GitHub APIs to search and retry for failed
// events.
func (s *Server) handleRetry() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.FromContext(ctx)

		if err := s.gcsLock.Acquire(ctx, s.lockTTL); err != nil {
			var lockErr *gcslock.LockHeldError
			if errors.As(err, &lockErr) {
				logger.Infof("lock is already acquired by another execution", "code", http.StatusOK, "body", errAcquireLock,
					"method", "RetrieveCheckpointID", "error", lockErr.Error())
			}
			// unable to obtain the lock, return a 200 so the scheduler doesnt attempt to reinvoke
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, http.StatusText(http.StatusOK))
			return
		}
		// read from checkpoint table
		cursor, err := s.datastore.RetrieveCheckpointID(ctx, s.checkpointTableID)
		if err != nil {
			logger.Errorw("failed to call RetrieveCheckpointID", "code", http.StatusInternalServerError, "body", errRetrieveCheckpoint,
				"method", "RetrieveCheckpointID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
			return
		}

		for cursor != "" {
			// call list deliveries API
			deliveries, res, err := s.github.ListDeliveries(ctx, &github.ListCursorOptions{Cursor: cursor})
			if err != nil {
				logger.Errorw("failed to call ListDeliveries", "code", http.StatusInternalServerError, "body", errCallingGitHub,
					"method", "RedeliverEvent", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
				return
			}

			// update the cursor
			cursor = res.Cursor

			// for each failed delivery, redeliver
			for i := 0; i < len(deliveries); i++ {
				event := deliveries[i]
				// check payload and see if its been successfully delivered, if so skip over it
				if *event.StatusCode == 200 {
					continue
				}

				// redeliver the failed event
				if err := s.github.RedeliverEvent(ctx, *event.ID); err != nil {
					logger.Errorw("failed to call RedeliverEvent", "code", http.StatusInternalServerError, "body", errCallingGitHub,
						"method", "RedeliverEvent", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
					return
				}
			}

			// periodic update to checkpoint
			if err := s.datastore.WriteCheckpointID(ctx, s.checkpointTableID, "checkpointID", time.Now().String()); err != nil {
				logger.Errorw("failed to call WriteCheckpointID", "code", http.StatusInternalServerError, "body", errWriteCheckpoint,
					"method", "RedeliverEvent", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
				return
			}
		}

		logger.Infow("successful", "code", http.StatusAccepted, "body", acceptedMessage)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, http.StatusText(http.StatusAccepted))
	})
}
