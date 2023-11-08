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

package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abcxyz/pkg/githubapp"
)

// ghTokenResponse is a go struct that maps to json structure of the response
// GitHub returns when requesting a token.
type ghTokenResponse struct {
	Token string `json:"token"`
}

// GitHubTokenSupplier is an interface defining a method to get a GitHub token.
type GitHubTokenSupplier interface {
	GitHubToken(ctx context.Context) (string, error)
}

// staticGitHubTokenSupplier is a GitHubTokenSupplier that simply holds a
// static token.
type staticGitHubTokenSupplier struct {
	token string
}

// GitHubToken returns the static token held by this struct.
func (g *staticGitHubTokenSupplier) GitHubToken(ctx context.Context) (string, error) {
	return g.token, nil
}

// gitHubAppTokenSupplier is a GitHubTokenSupplier that uses a GitHub App
// to request a token on demand.
type gitHubAppTokenSupplier struct {
	githubApp *githubapp.GitHubApp
}

// GitHubToken requests a token using the GitHub App represented by this struct.
func (g *gitHubAppTokenSupplier) GitHubToken(ctx context.Context) (string, error) {
	token, err := ReadAccessTokenForAllRepos(ctx, g.githubApp)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub access token: %w", err)
	}
	return token, nil
}

// NewStaticGitHubTokenSupplier creates a GitHubTokenSupplier that simply
// holds a static token and returns it whenever GitHubToken us called.
func NewStaticGitHubTokenSupplier(token string) GitHubTokenSupplier {
	return &staticGitHubTokenSupplier{
		token: token,
	}
}

// NewGitHubAppTokenSupplier creates a GitHubTokenSupplier that uses
// A GitHubApp to supply tokens.
func NewGitHubAppTokenSupplier(githubAppID, githubAppInstallationID string, githubAppPrivateKey *rsa.PrivateKey) GitHubTokenSupplier {
	githubAppConfig := githubapp.NewConfig(githubAppID, githubAppInstallationID, githubAppPrivateKey)
	return &gitHubAppTokenSupplier{
		githubApp: githubapp.New(githubAppConfig),
	}
}

// ReadAccessTokenForRepos generate a new access token with read permissions
// for the given repositories using the given GitHub App.
func ReadAccessTokenForRepos(ctx context.Context, githubApp *githubapp.GitHubApp, repositories ...string) (string, error) {
	tokenRequest := &githubapp.TokenRequest{
		Repositories: repositories,
		Permissions: map[string]string{
			"actions": "read",
		},
	}
	// @TODO(bradegler): This could use some caching. Requests to the same repos
	// can reuse a token without requesting a new one until it expires. Might be
	// better to implement that in pkg so that GitHubTokenMinter can take
	// advantage of it as well.
	ghAppJWT, err := githubApp.AccessToken(ctx, tokenRequest)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub access token: %w", err)
	}
	return parseJWT(ghAppJWT)
}

// ReadAccessTokenForAllRepos generate a new access token with read permissions
// for all repositories using the given GitHub app.
func ReadAccessTokenForAllRepos(ctx context.Context, githubApp *githubapp.GitHubApp) (string, error) {
	tokenRequest := &githubapp.TokenRequestAllRepos{
		Permissions: map[string]string{
			"actions": "read",
		},
	}
	ghAppJWT, err := githubApp.AccessTokenAllRepos(ctx, tokenRequest)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub access token: %w", err)
	}
	return parseJWT(ghAppJWT)
}

// parseJWT parses the given JWT and returns the embedded token.
func parseJWT(ghAppJWT string) (string, error) {
	// The token response is a json doc with a lot of information about the
	// token. All that is needed is the token itself.
	var ght ghTokenResponse
	if err := json.NewDecoder(strings.NewReader(ghAppJWT)).Decode(&ght); err != nil {
		return "", fmt.Errorf("failed to parse github token response: %w", err)
	}
	if ght.Token == "" {
		return "", fmt.Errorf("failed to parse github token response: no token in payload")
	}
	return ght.Token, nil
}
