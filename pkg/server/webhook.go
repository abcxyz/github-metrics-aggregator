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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pubsubpb "github.com/abcxyz/github-metrics-aggregator/protos/pubsub_schemas"
	"github.com/google/go-github/v48/github"
)

// event is the required pubsub topic schema for this application.

// processWebhookRequest handles the logic for receiving github webhooks and publishing to pubsub topic.
func (s *GitHubMetricsAggregatorServer) processWebhookRequest(r *http.Request) (int, string, error) {
	received := time.Now().UTC().Format(time.RFC3339Nano)
	deliveryID := strings.Join(r.Header[github.DeliveryIDHeader], " ")
	eventType := strings.Join(r.Header[github.EventTypeHeader], " ")
	signature := strings.Join(r.Header[github.SHA256SignatureHeader], " ")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, "Failed to read webhook payload.", fmt.Errorf("failed read webhook request body: %w", err)
	}

	if len(payload) == 0 {
		return http.StatusBadRequest, "No payload received.", fmt.Errorf("no payload received")
	}

	if err := github.ValidateSignature(signature, payload, []byte(s.webhookSecret)); err != nil {
		return http.StatusBadRequest, "Failed to validate webhook signature.", fmt.Errorf("failed to validate webhook payload: %w", err)
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
		return http.StatusInternalServerError, "Failed to create event JSON.", fmt.Errorf("failed to marshal event json: %w", err)
	}

	if err = s.messager.Send(context.Background(), eventBytes); err != nil {
		return http.StatusInternalServerError, "Failed to write to backend.", fmt.Errorf("failed to write messages: %w", err)
	}

	return http.StatusCreated, "Ok", nil
}
