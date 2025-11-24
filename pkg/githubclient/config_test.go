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
	"testing"

	"github.com/abcxyz/pkg/testutil"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "github_enterprise_server_url_wrong_format",
			cfg: &Config{
				GitHubEnterpriseServerURL: "test-url",
				GitHubAppID:               "test-github-app-id",
				GitHubPrivateKey:          "test-github-private-key",
			},
			wantErr: `GITHUB_ENTERPRISE_SERVER_URL does not start with "https://"`,
		},
		{
			name:    "missing_github_app_id",
			cfg:     &Config{},
			wantErr: `GITHUB_APP_ID is required`,
		},
		{
			name: "missing_github_private_key_and_kms_key_id",
			cfg: &Config{
				GitHubAppID: "test-github-app-id",
			},
			wantErr: `GITHUB_PRIVATE_KEY_SECRET or GITHUB_PRIVATE_KEY_KMS_KEY_ID is required`,
		},
		{
			name: "missing_github_private_key_kms_key_id",
			cfg: &Config{
				GitHubAppID:      "test-github-app-id",
				GitHubPrivateKey: "test-github-private-key",
			},
		},
		{
			name: "too_many_private_keys",
			cfg: &Config{
				GitHubAppID:              "test-github-app-id",
				GitHubPrivateKey:         "test-github-private-key",
				GitHubPrivateKeyKMSKeyID: "test-github-private-key-kms-key-id",
			},
			wantErr: `only one of GITHUB_PRIVATE_KEY_SECRET, GITHUB_PRIVATE_KEY_KMS_KEY_ID is required`,
		},
		{
			name: "success",
			cfg: &Config{
				GitHubAppID:      "test-github-app-id",
				GitHubPrivateKey: "test-github-private-key",
			},
		},
		{
			name: "success_with_enterprise_url",
			cfg: &Config{
				GitHubEnterpriseServerURL: "https://test-enterprise.com",
				GitHubAppID:               "test-github-app-id",
				GitHubPrivateKey:          "test-github-private-key",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			err := tc.cfg.Validate(ctx)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}
		})
	}
}
