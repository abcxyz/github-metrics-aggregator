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

	successMessage       = "Ok"
	errReadingPayload    = "Failed to read webhook payload."
	errNoPayload         = "No payload received."
	errInvalidSignature  = "Failed to validate webhook signature."
	errCreatingEventJSON = "Failed to create event JSON."
	errWritingToBackend  = "Failed to write to backend."
)

// handleWebhook handles the logic for receiving github webhooks and publishing
// to pubsub topic.
func (s *Server) handleWebhook() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.FromContext(r.Context())

		received := time.Now().UTC().Format(time.RFC3339Nano)
		deliveryID := r.Header.Get(DeliveryIDHeader)
		eventType := r.Header.Get(EventTypeHeader)
		signature := r.Header.Get(SHA256SignatureHeader)

		payload, err := io.ReadAll(io.LimitReader(r.Body, 25*mb))
		if err != nil {
			logger.Errorw("failed read webhook request body", "code", http.StatusInternalServerError, "body", errReadingPayload, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, errReadingPayload)
			return
		}

		if len(payload) == 0 {
			logger.Errorw("no payload received", "code", http.StatusBadRequest, "body", errNoPayload)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, errNoPayload)
			return
		}

		if !s.isValidSignature(signature, payload) {
			logger.Errorw("failed to validate webhook payload", "code", http.StatusUnauthorized, "body", errInvalidSignature, "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, errInvalidSignature)
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
			logger.Errorw("failed to marshal event json", "code", http.StatusInternalServerError, "body", errCreatingEventJSON, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, errCreatingEventJSON)
			return
		}

		if err = s.pubsub.Send(context.Background(), eventBytes); err != nil {
			logger.Errorw("failed to write messages to pubsub", "code", http.StatusInternalServerError, "body", errWritingToBackend, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, errWritingToBackend)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, successMessage)
	})
}

// isValidSignature validates the http request signature against the signature of the payload.
func (s *Server) isValidSignature(signature string, payload []byte) bool {
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(payload)
	got := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(signature), []byte(got)) == 1
}
