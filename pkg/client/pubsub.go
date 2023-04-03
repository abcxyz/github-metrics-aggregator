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

package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// PubSubMessenger implements the Messenger interface for Google Cloud pubsub.
type PubSubMessenger struct {
	projectID string
	topicID   string

	client *pubsub.Client
	topic  *pubsub.Topic
}

// NewPubSubMessenger creates a new instance of the PubSubMessenger.
func NewPubSubMessenger(ctx context.Context, projectID, topicID string, opts ...option.ClientOption) (*PubSubMessenger, error) {
	// pubsub client forces you to provide a projectID
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pubsub client: %w", err)
	}

	topic := client.Topic(topicID)

	return &PubSubMessenger{
		projectID: projectID,
		topicID:   topicID,
		client:    client,
		topic:     topic,
	}, nil
}

// Send sends a message to a Google Cloud pubsub topic.
func (p *PubSubMessenger) Send(ctx context.Context, msg []byte) error {
	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("pubsub: failed to get result: %w", err)
	}
	return nil
}

// Shutdown handles the graceful shutdown of the pubsub client.
func (p *PubSubMessenger) Shutdown() error {
	p.topic.Stop()
	if err := p.client.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}
	return nil
}
