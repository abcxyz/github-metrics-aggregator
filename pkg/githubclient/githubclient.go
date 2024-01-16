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
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/google/go-github/v56/github"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"golang.org/x/oauth2"

	"github.com/abcxyz/pkg/githubapp"
)

type GitHub struct {
	client *github.Client
}

// New creates a new instance of a GitHub client.
func New(ctx context.Context, appID, rsaPrivateKeyPEM string) (*GitHub, error) {
	// Read the private key.
	privateKey, err := readPrivateKey(rsaPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Intentionally sending an empty string for the installationID, it isn't used
	// when generating an app token.
	ghCfg := githubapp.NewConfig(appID, "", privateKey, githubapp.WithJWTTokenCaching(1*time.Minute))
	githubApp := githubapp.New(ghCfg)

	ts := oauth2.ReuseTokenSource(nil, githubApp)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

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

// readPrivateKey reads a RSA encrypted private key using PEM encoding as a string and returns an RSA key.
func readPrivateKey(rsaPrivateKeyPEM string) (*rsa.PrivateKey, error) {
	parsedKey, _, err := jwk.DecodePEM([]byte(rsaPrivateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
	}
	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
	}
	return privateKey, nil
}
