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
	"crypto"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/githubauth"
	"github.com/google/go-github/v61/github"
	"github.com/sethvargo/go-gcpkms/pkg/gcpkms"
)

type GitHub struct {
	client *github.Client
}

// New creates a new instance of a GitHub client.
func New(ctx context.Context, appID, rsaPrivateKeyPEM string) (*GitHub, error) {
	return NewGitHubEnterpriseClient(ctx, "", appID, rsaPrivateKeyPEM)
}

// NewGitHubEnterpriseClient creates a new instance of a GitHub client (for
// enterprise if enterpriseURL is non-empty).
func NewGitHubEnterpriseClient(ctx context.Context, enterpriseURL, appID, rsaPrivateKeyPEM string) (*GitHub, error) {
	app, err := NewGitHubApp(ctx, enterpriseURL, appID, rsaPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	ts := app.OAuthAppTokenSource()
	client, err := NewGitHubClient(ctx, ts, enterpriseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create github client: %w", err)
	}
	return &GitHub{client: client}, nil
}

func NewGitHubApp(ctx context.Context, enterpriseURL, appID, rsaPrivateKeyPEM string) (*githubauth.App, error) {
	signer, err := githubauth.NewPrivateKeySigner(rsaPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key signer: %w", err)
	}

	var appOpts []githubauth.Option
	if enterpriseURL != "" {
		appOpts = append(appOpts, githubauth.WithBaseURL(enterpriseURL+"/api/v3"))
	}

	app, err := githubauth.NewApp(appID, signer, appOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}
	return app, nil
}

// NewFromKMS creates a new GitHub client using a KMS key for authentication.
func NewFromKMS(ctx context.Context, signer *gcpkms.Signer, appID, enterpriseURL string) (*GitHub, error) {
	app, err := NewGitHubAppFromSigner(ctx, signer, enterpriseURL, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app from kms: %w", err)
	}

	ts := app.OAuthAppTokenSource()
	client, err := NewGitHubClient(ctx, ts, enterpriseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create github client: %w", err)
	}
	return &GitHub{client: client}, nil
}

func NewGitHubAppFromSigner(ctx context.Context, signer crypto.Signer, enterpriseURL, appID string) (*githubauth.App, error) {
	var appOpts []githubauth.Option
	if enterpriseURL != "" {
		appOpts = append(appOpts, githubauth.WithBaseURL(enterpriseURL+"/api/v3"))
	}

	app, err := githubauth.NewApp(appID, signer, appOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}
	return app, nil
}

func NewGitHubClient(ctx context.Context, ts oauth2.TokenSource, enterpriseURL string) (*github.Client, error) {
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if enterpriseURL == "" {
		return client, nil
	}

	client, err := client.WithEnterpriseURLs(enterpriseURL, enterpriseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create enterprise client: %w", err)
	}
	return client, nil
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
