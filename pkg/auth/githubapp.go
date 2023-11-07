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
	"encoding/json"
	"fmt"

	"github.com/abcxyz/pkg/githubapp"
)

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

// parseJWT parses the give JWT and returns the embedded token.
func parseJWT(ghAppJWT string) (string, error) {
	// The token response is a json doc with a lot of information about the
	// token. All that is needed is the token itself.
	var tokenResp map[string]any
	if err := json.Unmarshal([]byte(ghAppJWT), &tokenResp); err != nil {
		return "", fmt.Errorf("malformed GitHub token response: %w", err)
	}
	t, ok := tokenResp["token"]
	if !ok {
		return "", fmt.Errorf("malformed GitHub token response - missing token attribute")
	}
	token, ok := t.(string)
	if !ok {
		return "", fmt.Errorf("malformed GitHub token: wanted string got %T", token)
	}
	fmt.Printf("%s\n", tokenResp)
	return token, nil
}
