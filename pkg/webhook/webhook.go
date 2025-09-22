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

package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	pubsubpb "github.com/abcxyz/github-metrics-aggregator/protos/pubsub_schemas"
	"github.com/abcxyz/pkg/logging"
)

const (
	// SHA256SignatureHeader is the GitHub header key used to pass the HMAC-SHA256 hexdigest.
	SHA256SignatureHeader = "X-Hub-Signature-256"

	// EventTypeHeader is the GitHub header key used to pass the event type.
	EventTypeHeader = "X-Github-Event"

	// DeliveryIDHeader is the GitHub header key used to pass the unique ID for the webhook event.
	DeliveryIDHeader = "X-Github-Delivery"

	// mb is used for conversion to megabytes.
	mb = 1000000
)

var (
	statusOK = map[string]string{"status": "ok"}

	errReadingPayload    = fmt.Errorf("failed to read webhook payload")
	errNoPayload         = fmt.Errorf("no payload received")
	errInvalidSignature  = fmt.Errorf("failed to validate webhook signature")
	errCreatingEventJSON = fmt.Errorf("failed to create event json")
	errWritingToBackend  = fmt.Errorf("failed to write to backend")
)

// handleWebhook handles the logic for receiving github webhooks and publishing
// to pubsub topic.
func (s *Server) handleWebhook() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		ctx := r.Context()
		logger := logging.FromContext(ctx)

		received := now.Format(time.RFC3339Nano)
		deliveryID := r.Header.Get(DeliveryIDHeader)
		eventType := r.Header.Get(EventTypeHeader)
		signature := r.Header.Get(SHA256SignatureHeader)

		payload, err := io.ReadAll(io.LimitReader(r.Body, 25*mb))
		if err != nil {
			logger.ErrorContext(ctx, "failed read webhook request body",
				"code", http.StatusInternalServerError,
				"body", errReadingPayload,
				"error", err)
			s.h.RenderJSON(w, http.StatusInternalServerError, errReadingPayload)
			return
		}

		if len(payload) == 0 {
			logger.ErrorContext(ctx, "no payload received",
				"code", http.StatusBadRequest,
				"body", errNoPayload)
			s.h.RenderJSON(w, http.StatusBadRequest, errNoPayload)
			return
		}

		if !s.isValidSignature(signature, payload) {
			logger.ErrorContext(ctx, "failed to validate webhook payload",
				"code", http.StatusUnauthorized,
				"body", errInvalidSignature,
				"error", err)
			s.h.RenderJSON(w, http.StatusUnauthorized, errInvalidSignature)
			return
		}

		event := &pubsubpb.Event{
			Received:   received,
			DeliveryId: deliveryID,
			Signature:  signature,
			Event:      eventType,
			Payload:    string(payload),
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			logger.ErrorContext(ctx, "failed to marshal event json",
				"code", http.StatusInternalServerError,
				"body", errCreatingEventJSON,
				"error", err)
			s.h.RenderJSON(w, http.StatusInternalServerError, errCreatingEventJSON)
			return
		}

		if err := s.eventsPubsub.Send(context.Background(), eventBytes); err != nil {
			logger.ErrorContext(ctx, "failed to write messages to event pubsub",
				"code", http.StatusInternalServerError,
				"body", errWritingToBackend,
				"error", err)

			exceeds, bqQueryErr := s.datastore.
				FailureEventsExceedsRetryLimit(ctx, s.failureEventTableID, deliveryID, s.retryLimit)
			if bqQueryErr != nil {
				logger.ErrorContext(ctx, "failed to call BigQuery",
					"method", "FailureEventsExceedsRetryLimit",
					"code", http.StatusInternalServerError,
					"body", errWritingToBackend,
					"error", bqQueryErr)
			} else if exceeds {
				// exceeds the limit, write to DLQ
				if err := s.dlqEventsPubsub.Send(context.Background(), eventBytes); err != nil {
					logger.ErrorContext(ctx, "failed to write messages to pubsub DLQ",
						"method", "SendDLQ",
						"code", http.StatusInternalServerError,
						"body", errWritingToBackend,
						"error", err)

					// potential outage with PubSub, fail this iteration so an additional
					// attempt can be made in the future
					s.h.RenderJSON(w, http.StatusInternalServerError, errWritingToBackend)
					return
				}

				// return a 200 so GitHub doesn't report a failed delivery
				s.h.RenderJSON(w, http.StatusCreated, statusOK)
				return
			} else {
				// record an entry in the failure events table
				if err := s.datastore.
					WriteFailureEvent(ctx, s.failureEventTableID, deliveryID, now.Format(time.DateTime)); err != nil {
					logger.ErrorContext(ctx, "failed to call BigQuery",
						"method", "WriteFailureEvent",
						"code", http.StatusInternalServerError,
						"body", errWritingToBackend,
						"error", err)
				}
			}

			s.h.RenderJSON(w, http.StatusInternalServerError, errWritingToBackend)
			return
		}

		s.h.RenderJSON(w, http.StatusCreated, statusOK)
	})
}

// isValidSignature validates the http request signature against the signature of the payload.
func (s *Server) isValidSignature(signature string, payload []byte) bool {
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(payload)
	got := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(signature), []byte(got)) == 1
}
