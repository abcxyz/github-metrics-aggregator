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

// Package relay contains the HTTP handler and logic for the relay service.
package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	protos "github.com/abcxyz/github-metrics-aggregator/protos/pubsub_schemas"
	"github.com/abcxyz/pkg/logging"
)

const mb = 1 << 20

type wrappedMessage struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
		ID         string            `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func (s *Server) handleRelay() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.FromContext(ctx)

		defer r.Body.Close()

		req, err := io.ReadAll(io.LimitReader(r.Body, 25*mb))
		if err != nil {
			logger.ErrorContext(ctx, "failed to read request body", "error", err)
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		var m wrappedMessage
		if err := json.Unmarshal(req, &m); err != nil {
			logger.ErrorContext(ctx, "failed to decode request body", "error", err)
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		logger.InfoContext(ctx, "received pubsub message",
			"subscription", m.Subscription,
			"message_id", m.Message.ID,
			"data", string(m.Message.Data),
			"attributes", m.Message.Attributes,
		)

		var event protos.Event
		if err := json.Unmarshal(m.Message.Data, &event); err != nil {
			logger.ErrorContext(ctx, "failed to decode request event", "error", err)
			http.Error(w, "failed to decode request event", http.StatusBadRequest)
			return
		}

		enrichedEvent := &protos.EnrichedEvent{
			DeliveryId: event.GetDeliveryId(),
			Signature:  event.GetSignature(),
			Received:   event.GetReceived(),
			Event:      event.GetEvent(),
			Payload:    event.GetPayload(),
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.GetPayload()), &payload); err != nil {
			logger.ErrorContext(ctx, "failed to decode event payload", "error", err)
			http.Error(w, "failed to decode event payload", http.StatusBadRequest)
			return
		}

		if enterprise, ok := payload["enterprise"].(map[string]interface{}); ok {
			if id, ok := enterprise["id"].(float64); ok {
				enrichedEvent.EnterpriseId = strconv.Itoa(int(id))
			}
			if name, ok := enterprise["name"].(string); ok {
				enrichedEvent.EnterpriseName = name
			}
		}
		if organization, ok := payload["organization"].(map[string]interface{}); ok {
			if id, ok := organization["id"].(float64); ok {
				enrichedEvent.OrganizationId = strconv.Itoa(int(id))
			}
			if login, ok := organization["login"].(string); ok {
				enrichedEvent.OrganizationName = login
			}
		}
		if repository, ok := payload["repository"].(map[string]interface{}); ok {
			if id, ok := repository["id"].(float64); ok {
				enrichedEvent.RepositoryId = strconv.Itoa(int(id))
			}
			if fullName, ok := repository["full_name"].(string); ok {
				enrichedEvent.RepositoryName = fullName
			}
		}

		attrs := map[string]string{
			"enterprise_id":     enrichedEvent.GetEnterpriseId(),
			"enterprise_name":   enrichedEvent.GetEnterpriseName(),
			"organization_id":   enrichedEvent.GetOrganizationId(),
			"organization_name": enrichedEvent.GetOrganizationName(),
			"repository_id":     enrichedEvent.GetRepositoryId(),
			"repository_name":   enrichedEvent.GetRepositoryName(),
			"event":             enrichedEvent.GetEvent(),
		}

		data, err := json.Marshal(enrichedEvent)
		if err != nil {
			logger.ErrorContext(ctx, "failed to encode enriched event", "error", err)
			http.Error(w, "failed to encode enriched event", http.StatusInternalServerError)
			return
		}

		if err := s.relayMessenger.Send(ctx, data, attrs); err != nil {
			logger.ErrorContext(ctx, "failed to write message to relay topic",
				"code", http.StatusInternalServerError,
				"body", "error writing message to relay topic",
				"error", err)
		}

		w.WriteHeader(http.StatusOK)
	})
}
