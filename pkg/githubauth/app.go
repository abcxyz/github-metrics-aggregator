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

package githubauth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/abcxyz/pkg/githubapp"
)

// AppScope is an interface which determines the repositories scope for which
// the app should request tokens.
type AppScope interface {
	*allReposScope | *selectedReposScope
}

type allReposScope struct{}

// ForAllRepos creates a scope that requests permissions on all repositories.
func ForAllRepos() *allReposScope {
	return &allReposScope{}
}

type selectedReposScope struct {
	repos []string
}

// ForRepos creates a scope that requests permissions on the given repositories.
func ForRepos(r ...string) *selectedReposScope {
	return &selectedReposScope{
		repos: r,
	}
}

var _ TokenSource = (*AppTokenSource[*allReposScope])(nil)

// AppTokenSource is a GitHubToken provider that authenticates as a GitHub App.
type AppTokenSource[T AppScope] struct {
	app   *githubapp.GitHubApp
	scope T
}

// NewAppTokenSource returns a [AppTokenSource] which authenticates as a GitHub
// App and returns a GitHub token.
func NewAppTokenSource[T AppScope](appID, installationID, privateKeyPEM string, scope T) (*AppTokenSource[T], error) {
	var merr error
	if appID == "" {
		merr = errors.Join(merr, fmt.Errorf("missing appID"))
	}
	if installationID == "" {
		merr = errors.Join(merr, fmt.Errorf("missing installationID"))
	}
	if privateKeyPEM == "" {
		merr = errors.Join(merr, fmt.Errorf("missing privateKeyPEM"))
	}
	if merr != nil {
		return nil, fmt.Errorf("failed to create app token source: %w", merr)
	}

	privateKey, err := parseRSAPrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create app token source: %w", err)
	}

	app := githubapp.New(githubapp.NewConfig(appID, installationID, privateKey))

	return &AppTokenSource[T]{
		app:   app,
		scope: scope,
	}, nil
}

func parseRSAPrivateKeyPEM(data string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, fmt.Errorf("failed to parse pem: no pem block found")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key pem: %w", err)
	}
	return key, nil
}

// GitHubToken implements [TokenSource].
func (s *AppTokenSource[T]) GitHubToken(ctx context.Context) (string, error) {
	switch t := any(s.scope).(type) {
	case *allReposScope:
		resp, err := s.app.AccessTokenAllRepos(ctx, &githubapp.TokenRequestAllRepos{
			Permissions: map[string]string{
				"actions":       "read",
				"pull-requests": "read",
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to get github access token for all repos: %w", err)
		}
		return parseAppTokenResponse(resp)

	case *selectedReposScope:
		resp, err := s.app.AccessToken(ctx, &githubapp.TokenRequest{
			Repositories: t.repos,
			Permissions: map[string]string{
				"actions":       "read",
				"pull-requests": "read",
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to get github access token for repos %q: %w", t.repos, err)
		}
		return parseAppTokenResponse(resp)

	default:
		panic("impossible")
	}
}

// GitHubApp returns the underlying GitHubApp.
func (s *AppTokenSource[T]) GitHubApp() *githubapp.GitHubApp {
	return s.app
}

type appTokenResponse struct {
	Token string `json:"token"`
}

// parseAppTokenResponse parses the given JWT and returns the embedded token.
func parseAppTokenResponse(data string) (string, error) {
	var resp appTokenResponse
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&resp); err != nil {
		return "", fmt.Errorf("failed to parse json: %w", err)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("no token in json response")
	}
	return resp.Token, nil
}
