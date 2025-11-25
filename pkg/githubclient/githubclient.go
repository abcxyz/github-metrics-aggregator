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

// Package githubclient is a wrapper around the GitHub App for common
// operations.
package githubclient

import (
	"context"
	"crypto"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/google/go-github/v61/github"
	"github.com/sethvargo/go-gcpkms/pkg/gcpkms"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/githubauth"
)

// Client is a wrapper around a GitHub HTTP client and an authenticated GitHub
// App.
type Client struct {
	config       *Config
	app          *githubauth.App
	githubClient *github.Client
}

// New creates a new [Client] from the given config.
func New(ctx context.Context, c *Config) (*Client, error) {
	var signer crypto.Signer
	var err error

	if c.GitHubPrivateKeyKMSKeyID != "" {
		client, err := kms.NewKeyManagementClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create new key management client: %w", err)
		}

		signer, err = gcpkms.NewSigner(ctx, client, c.GitHubPrivateKeyKMSKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to create app signer: %w", err)
		}
	} else if c.GitHubPrivateKeySecretID != "" {
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create secretmanager client: %w", err)
		}
		defer client.Close()

		// Build the request.
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: c.GitHubPrivateKeySecretID,
		}

		// Call the API.
		result, err := client.AccessSecretVersion(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to access secret version: %w", err)
		}

		signer, err = githubauth.NewPrivateKeySigner(string(result.GetPayload().GetData()))
		if err != nil {
			return nil, fmt.Errorf("failed to create private key signer: %w", err)
		}
	} else if c.GitHubPrivateKey != "" {
		signer, err = githubauth.NewPrivateKeySigner(c.GitHubPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create private key signer: %w", err)
		}
	}

	// Create the GitHub App from our shared library
	var appOpts []githubauth.Option
	if v := c.GitHubEnterpriseServerURL; v != "" {
		appOpts = append(appOpts, githubauth.WithBaseURL(v+"/api/v3"))
	}
	app, err := githubauth.NewApp(c.GitHubAppID, signer, appOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	// Create the authenticated github API client
	githubClient := github.NewClient(oauth2.NewClient(ctx, app.OAuthAppTokenSource()))
	if v := c.GitHubEnterpriseServerURL; v != "" {
		var err error
		githubClient, err = githubClient.WithEnterpriseURLs(v, v)
		if err != nil {
			return nil, fmt.Errorf("failed to create enterprise client: %w", err)
		}
	}

	return &Client{
		config:       c,
		app:          app,
		githubClient: githubClient,
	}, nil
}

// App returns the underlying [githubauth.App].
func (c *Client) App() *githubauth.App {
	return c.app
}

// GitHubClientFromTokenSource creates a new GitHub client from the given token
// source. It inherits any configuration from the GitHub config (like enterprise
// URL).
func (c *Client) GitHubClientFromTokenSource(ctx context.Context, ts oauth2.TokenSource) (*github.Client, error) {
	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))
	if v := c.config.GitHubEnterpriseServerURL; v != "" {
		var err error
		githubClient, err = githubClient.WithEnterpriseURLs(v, v)
		if err != nil {
			return nil, fmt.Errorf("failed to create enterprise client: %w", err)
		}
	}
	return githubClient, nil
}

// ListDeliveries lists a paginated result of event deliveries.
func (c *Client) ListDeliveries(ctx context.Context, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error) {
	deliveries, resp, err := c.githubClient.Apps.ListHookDeliveries(ctx, opts)
	if err != nil {
		return deliveries, resp, fmt.Errorf("failed to list deliveries: %w", err)
	}
	return deliveries, resp, nil
}

// RedeliverEvent redelivers a failed event which will be picked up by the webhook service.
func (c *Client) RedeliverEvent(ctx context.Context, deliveryID int64) error {
	_, _, err := c.githubClient.Apps.RedeliverHookDelivery(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to redeliver event: %w", err)
	}
	return nil
}
