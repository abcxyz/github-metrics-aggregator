// Copyright 2026 The Authors (see AUTHORS file)
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

package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/abcxyz/github-metrics-aggregator/pkg/events"
	"github.com/abcxyz/pkg/logging"
)

type MessageEnricher interface {
	Enrich(ctx context.Context, req []byte) ([]byte, map[string]string, error)
}

func NewDefaultMessageEnricher() MessageEnricher {
	return &defaultMessageEnricher{}
}

type defaultMessageEnricher struct{}

func (d *defaultMessageEnricher) Enrich(ctx context.Context, req []byte) ([]byte, map[string]string, error) {
	logger := logging.FromContext(ctx)
	var m wrappedMessage
	if err := json.Unmarshal(req, &m); err != nil {
		logger.ErrorContext(ctx, "failed to decode request body", "error", err)
		return nil, nil, fmt.Errorf("failed to decode request body: %w", err)
	}

	logger.DebugContext(ctx, "received pubsub message",
		"subscription", m.Subscription,
		"message_id", m.Message.ID,
		"data", string(m.Message.Data),
		"attributes", m.Message.Attributes,
	)

	var event events.Event
	if err := json.Unmarshal(m.Message.Data, &event); err != nil {
		logger.ErrorContext(ctx, "failed to decode request event", "error", err)
		return nil, nil, fmt.Errorf("failed to decode request event: %w", err)
	}

	enrichedEvent := &events.EnrichedEvent{
		DeliveryId: event.DeliveryId,
		Signature:  event.Signature,
		Received:   event.Received,
		Event:      event.Event,
		Payload:    event.Payload,
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		logger.ErrorContext(ctx, "failed to decode event payload", "error", err)
		return nil, nil, fmt.Errorf("failed to decode event payload: %w", err)
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
		"enterprise_id":     enrichedEvent.EnterpriseId,
		"enterprise_name":   enrichedEvent.EnterpriseName,
		"organization_id":   enrichedEvent.OrganizationId,
		"organization_name": enrichedEvent.OrganizationName,
		"repository_id":     enrichedEvent.RepositoryId,
		"repository_name":   enrichedEvent.RepositoryName,
		"event":             enrichedEvent.Event,
	}

	data, err := json.Marshal(enrichedEvent)
	if err != nil {
		logger.ErrorContext(ctx, "failed to encode enriched event", "error", err)
		return nil, nil, fmt.Errorf("failed to encode enriched event: %w", err)
	}
	return data, attrs, nil
}
