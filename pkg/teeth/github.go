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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/google/go-github/v56/github"
	"github.com/sethvargo/go-retry"
)

var (
	GitHubErrRateLimit          github.RateLimitError
	GitHubErrSecondaryRateLimit github.AbuseRateLimitError
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

// NewGitHub initializes the GitHub client with config and authentication.
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

// CreateInvocationCommentWithRetry makes a call to make an invocation comment
// to the PR by number.
//
// It leverages a default retry policy to get around GitHub's API rate limits.
func (gh *GitHub) CreateInvocationCommentWithRetry(ctx context.Context, prNum int, invocationComment string) error {
	backoff := retryFunc(retryMinWaitDuration)
	backoff = retry.WithMaxRetries(retryMaxAttempts, backoff)
	if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		var err error
		resp, err := gh.commentPRFunc(ctx, gh.config.Owner, gh.config.Repo, prNum, invocationComment)

		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}

		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}

		if shouldRetry(statusCode, err) {
			// Just in case err is nil but status code is not 2XX
			return retry.RetryableError(fmt.Errorf("retrying error with status %d: %w", statusCode, err))
		}

		if err != nil {
			return fmt.Errorf("non-retryable error response: %w", err)
		}

		if statusCode == http.StatusNotFound || statusCode == http.StatusGone {
			return fmt.Errorf("resource does not exist for %s/%s/pull/%d", gh.config.Owner, gh.config.Repo, prNum)
		}
		// Read the entire response.
		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2mb
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		bodyStr := string(bytes.TrimSpace(body))

		if statusCode != http.StatusCreated {
			return fmt.Errorf("non-201 response: %v", bodyStr)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create invocation comment: %w", err)
	}

	return nil
}

func shouldRetry(statusCode int, err error) bool {
	// Retry if rate-limited. See https://github.com/google/go-github#rate-limiting.
	if GitHubErrRateLimit.Is(err) || GitHubErrSecondaryRateLimit.Is(err) {
		return true
	}

	// See list of possible status codes in https://docs.github.com/en/rest/issues/comments
	if statusCode == http.StatusForbidden {
		return true
	}
	if statusCode == http.StatusUnprocessableEntity {
		return true
	}

	// Retry server-side errors.
	if statusCode == http.StatusInternalServerError {
		return true
	}

	return false
}
