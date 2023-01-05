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

package messaging

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"
)

// PubSubMessager implements the Messager interface for Google Cloud pubsub.
type PubSubMessager struct {
	projectID string
	topicID   string

	client *pubsub.Client
	topic  *pubsub.Topic
}

// NewPubSubMessager creates a new instance of the PubSubMessager.
func NewPubSubMessager(ctx context.Context, projectID, topicID string, logger *zap.SugaredLogger) (*PubSubMessager, error) {
	// pubsub client forces you to provide a projectID
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pubsub client: %w", err)
	}

	topic := client.Topic(topicID)

	return &PubSubMessager{
		projectID: projectID,
		topicID:   topicID,
		client:    client,
		topic:     topic,
	}, nil
}

// Send sends a message to a Google Cloud pubsub topic.
func (p *PubSubMessager) Send(ctx context.Context, msg []byte) error {
	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	return nil
}

// Cleanup handles the graceful shutdown of the pubsub client.
func (p *PubSubMessager) Cleanup(ctx context.Context) error {
	p.topic.Stop()
	if err := p.client.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}
	return nil
}
