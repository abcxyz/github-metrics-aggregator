// Copyright 2022 GitHub Metrics Aggregator authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/google/go-github/v48/github"
)

// Event is the required pubsub topic schema for this application.
type Event struct {
	DeliveryID string `json:"delivery_id"`
	Signature  string `json:"signature"`
	Received   string `json:"received"`
	Event      string `json:"event"`
	Payload    string `json:"payload"`
}

// HandleWebhook handles the logic for receiving github webhooks and publishing to pubsub topic.
func (s *Server) HandleWebhook(w http.ResponseWriter, req *http.Request) {
	received := time.Now().UTC().Format(time.RFC3339Nano)
	deliveryID := strings.Join(req.Header[github.DeliveryIDHeader], " ")
	eventType := strings.Join(req.Header[github.EventTypeHeader], " ")
	signature := strings.Join(req.Header[github.SHA256SignatureHeader], " ")

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		s.logger.Errorf("failed read webhook request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to read webhook payload.")
		return
	}

	if len(payload) == 0 {
		s.logger.Error("no payload received")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No payload received.")
		return
	}

	if err := github.ValidateSignature(signature, payload, []byte(s.config.WebhookSecret)); err != nil {
		s.logger.Errorf("failed to validate webhook payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to validate webhook signature.")
		return
	}

	event := &Event{
		Received:   received,
		DeliveryID: deliveryID,
		Signature:  signature,
		Event:      eventType,
		Payload:    string(payload),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		s.logger.Errorf("failed to marshal event json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to create event JSON.")
		return
	}

	err = s.messager.Send(context.Background(), eventBytes)
	if err != nil {
		s.logger.Errorf("failed to write messages: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to write to backend.")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
