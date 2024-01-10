// Copyright 2024 The Authors (see AUTHORS file)
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

package teeth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/google/go-github/v56/github"
	"github.com/sethvargo/go-retry"
)

const ()

var (
	// can be overriden for testing
	retryMinWaitDuration        = 1 * time.Second
	retryMaxAttempts     uint64 = 4
	retryFunc                   = retry.NewFibonacci
)

// GHConfig contains configuration parameters for the GitHub client.
type GHConfig struct {
	Owner string
	Repo  string
}

// GitHub is just a wrapper around githubclient.GitHub for functionality specific to teeth.
// Use NewGitHub to initialize this.
type GitHub struct {
	*githubclient.GitHub

	config *GHConfig
	// can be overriden for testing
	commentPRFunc func(ctx context.Context, owner, repo string, prNumber int, comment string) (*github.Response, error)
}

func NewGitHub(ctx context.Context, config *GHConfig, appID, rsaPrivateKeyPEM string) (*GitHub, error) {
	client, err := githubclient.New(ctx, appID, rsaPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to setup github client: %w", err)
	}
	return &GitHub{
		GitHub:        client,
		config:        config,
		commentPRFunc: client.CommentPR,
	}, nil
}

func convertRetryable(statusCode int) error {
	// See list of status codes in https://docs.github.com/en/rest/issues/comments
	switch statusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusForbidden, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		return retry.RetryableError(fmt.Errorf("status code %v is marked as retryable", statusCode))
	}
	return fmt.Errorf("response has non-retryable error with status %v", statusCode)
}

// CreateInvocationCommentWithRetry makes a call to make an invocation comment
// to the PR by number.
//
// It leverages a default retry policy to get around GitHub's API rate limits.
func (gh *GitHub) CreateInvocationCommentWithRetry(ctx context.Context, prNum int, invocationComment string) error {
	backoff := retryFunc(retryMinWaitDuration)
	backoff = retry.WithMaxRetries(retryMaxAttempts, backoff)
	if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		resp, err := gh.commentPRFunc(ctx, gh.config.Owner, gh.config.Repo, prNum, invocationComment)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return convertRetryable(resp.StatusCode)
	}); err != nil {
		return fmt.Errorf("failed to create invocation comment: %w", err)
	}
	return nil
}
