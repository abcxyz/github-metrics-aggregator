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

package githubclient

import (
	"context"
	"fmt"

	"github.com/google/go-github/v61/github"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/githubauth"
)

type GitHub struct {
	client *github.Client
}

// New creates a new instance of a GitHub client.
func New(ctx context.Context, appID, rsaPrivateKeyPEM string) (*GitHub, error) {
	app, err := githubauth.NewApp(appID, rsaPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	ts := app.OAuthAppTokenSource()
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	return &GitHub{
		client: client,
	}, nil
}

// ListDeliveries lists a paginated result of event deliveries.
func (gh *GitHub) ListDeliveries(ctx context.Context, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error) {
	deliveries, resp, err := gh.client.Apps.ListHookDeliveries(ctx, opts)
	if err != nil {
		return deliveries, resp, fmt.Errorf("failed to list deliveries: %w", err)
	}
	return deliveries, resp, nil
}

// RedeliverEvent redelivers a failed event which will be picked up by the webhook service.
func (gh *GitHub) RedeliverEvent(ctx context.Context, deliveryID int64) error {
	_, _, err := gh.client.Apps.RedeliverHookDelivery(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to redeliver event: %w", err)
	}
	return nil
}
