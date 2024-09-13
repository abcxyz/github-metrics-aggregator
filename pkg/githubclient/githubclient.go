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

func NewWithPermissions(ctx context.Context, appID, gitHubInstallID, rsaPrivateKeyPEM string, permissions map[string]string, authOpts ...githubauth.Option) (*GitHub, error) {
	app, err := githubauth.NewApp(appID, rsaPrivateKeyPEM, authOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	installation, err := app.InstallationForID(ctx, gitHubInstallID)
	if err != nil {
		return nil, fmt.Errorf("failed to get github app installation: %w", err)
	}

	ts := installation.AllReposOAuth2TokenSource(ctx, permissions)

	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &GitHub{
		client: client,
	}, nil
}

func (gh *GitHub) WithBaseURL(url string) (*GitHub, error) {
	client, err := gh.client.WithEnterpriseURLs(url, url)
	if err != nil {
		return nil, fmt.Errorf("error making github client with enterprise URLs: %w", err)
	}
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

// CommentPR creates a PR comment.
func (gh *GitHub) CommentPR(ctx context.Context, owner, repo string, prNumber int, comment string) (*github.Response, error) {
	_, resp, err := gh.client.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{
		Body: github.String(comment),
	})
	if err != nil {
		return nil, fmt.Errorf("could not call GitHub issues create comment API: %w", err)
	}
	return resp, nil
}

func (gh *GitHub) DoRequest(ctx context.Context, method, urlStr string, body interface{}, opts ...github.RequestOption) (*github.Response, error) {
	req, err := gh.client.NewRequest(method, urlStr, body, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub request %s %s: %w", method, urlStr, err)
	}
	resp, err := gh.client.BareDo(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("error calling GitHub client with request %s %s: %w", method, urlStr, err)
	}
	return resp, nil
}
