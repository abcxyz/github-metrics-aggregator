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
			DeliveryId: event.DeliveryId,
			Signature:  event.Signature,
			Received:   event.Received,
			Event:      event.Event,
			// Payload:    event.Payload,
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			logger.ErrorContext(ctx, "failed to decode event payload", "error", err)
			http.Error(w, "failed to decode event payload", http.StatusBadRequest)
			return
		}

		if enterprise, ok := payload["enterprise"].(map[string]interface{}); ok {
			enrichedEvent.EnterpriseId = strconv.Itoa(int(enterprise["id"].(float64)))
			enrichedEvent.EnterpriseName = enterprise["name"].(string)
		}
		if organization, ok := payload["organization"].(map[string]interface{}); ok {
			enrichedEvent.OrganizationId = strconv.Itoa(int(organization["id"].(float64)))
			enrichedEvent.OrganizationName = organization["login"].(string)
		}
		if repository, ok := payload["repository"].(map[string]interface{}); ok {
			enrichedEvent.RepositoryId = strconv.Itoa(int(repository["id"].(float64)))
			enrichedEvent.RepositoryName = repository["full_name"].(string)
		}

		data, err := json.Marshal(enrichedEvent)
		if err != nil {
			logger.ErrorContext(ctx, "failed to encode enriched event", "error", err)
			http.Error(w, "failed to encode enriched event", http.StatusInternalServerError)
			return
		}

		if err := s.relayMessenger.Send(ctx, data); err != nil {
			logger.ErrorContext(ctx, "failed to write message to relay topic",
				"code", http.StatusInternalServerError,
				"body", "error writing message to relay topic",
				"error", err)
		}

		w.WriteHeader(http.StatusOK)
	})
}
