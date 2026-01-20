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

package pubsub

import (
	"testing"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewPubSubMessenger_SetsTimeout(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	projectID := "test-project"
	topicID := "test-topic"
	timeout := 15 * time.Second

	// Use WithoutAuthentication and a dummy endpoint to avoid connection attempts
	messenger, err := NewPubSubMessenger(ctx, projectID, topicID, timeout,
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithEndpoint("localhost:9090"), // Dummy endpoint
	)
	if err != nil {
		t.Fatalf("NewPubSubMessenger failed: %v", err)
	}

	p, ok := messenger.(*PubSubMessenger)
	if !ok {
		t.Fatalf("messenger is not *PubSubMessenger")
	}

	if p.projectID != projectID {
		t.Errorf("expected projectID %q, got %q", projectID, p.projectID)
	}

	if p.topicID != topicID {
		t.Errorf("expected topicID %q, got %q", topicID, p.topicID)
	}

	if p.topic.PublishSettings.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, p.topic.PublishSettings.Timeout)
	}
}
