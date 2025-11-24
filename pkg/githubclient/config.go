// Copyright 2025 The Authors (see AUTHORS file)
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
	"errors"
	"fmt"
	"strings"

	"github.com/abcxyz/pkg/cli"
)

// Config represents the shared GitHub App configuration.
type Config struct {
	// GitHubEnterpriseServerURL tis he GitHub Enterprise Server instance URL, in
	// the format "https://[hostname]".
	GitHubEnterpriseServerURL string

	// GitHubAppID is the GitHub App ID. This comes from the GitHub API.
	GitHubAppID string

	// GitHubPrivateKey is the GitHub App private key.
	GitHubPrivateKey string

	// GitHubPrivateKeyKMSKeyID is the KMS key ID to use for the GitHub App
	// private key. This is mutually-exclusive with [GitHubPrivateKey].
	GitHubPrivateKeyKMSKeyID string
}

// Validate does sanity checking on the configuration.
func (c *Config) Validate(ctx context.Context) error {
	var merr error
	if c.GitHubEnterpriseServerURL != "" && !strings.HasPrefix(c.GitHubEnterpriseServerURL, "https://") {
		merr = errors.Join(merr, fmt.Errorf(`GITHUB_ENTERPRISE_SERVER_URL does not start with "https://"`))
	}

	if c.GitHubAppID == "" {
		merr = errors.Join(merr, fmt.Errorf("GITHUB_APP_ID is required"))
	}

	if c.GitHubPrivateKey == "" && c.GitHubPrivateKeyKMSKeyID == "" {
		merr = errors.Join(merr, fmt.Errorf("GITHUB_PRIVATE_KEY_SECRET or GITHUB_PRIVATE_KEY_KMS_KEY_ID is required"))
	}

	if c.GitHubPrivateKey != "" && c.GitHubPrivateKeyKMSKeyID != "" {
		merr = errors.Join(merr, fmt.Errorf("only one of GITHUB_PRIVATE_KEY_SECRET, GITHUB_PRIVATE_KEY_KMS_KEY_ID is required"))
	}

	return merr
}

// ToFlags registers the GitHub flags.
func (c *Config) ToFlags(set *cli.FlagSet) {
	f := set.NewSection("GITHUB OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:   "github-enterprise-server_url",
		Target: &c.GitHubEnterpriseServerURL,
		EnvVar: "GITHUB_ENTERPRISE_SERVER_URL",
		Usage:  `The GitHub Enterprise Server instance URL, format "http(s)://[hostname]"`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-id",
		Target: &c.GitHubAppID,
		EnvVar: "GITHUB_APP_ID",
		Usage:  `The provisioned GitHub App ID.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key",
		Target: &c.GitHubPrivateKey,
		EnvVar: "GITHUB_PRIVATE_KEY_SECRET",
		Usage:  `The GitHub App private key. This is typically sourced from a secret manager via the GITHUB_PRIVATE_KEY_SECRET environment variable.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-private-key-kms-key-id",
		Target: &c.GitHubPrivateKeyKMSKeyID,
		EnvVar: "GITHUB_PRIVATE_KEY_KMS_KEY_ID",
		Usage:  `The KMS key ID for the GitHub App private key.`,
	})
}
