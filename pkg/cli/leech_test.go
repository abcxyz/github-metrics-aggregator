// Copyright 2023 Google LLC
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

package cli

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"testing"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
	"github.com/sethvargo/go-envconfig"
)

func TestLeechCommand(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))
	testPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemPrivateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(testPrivateKey),
	}
	keyBytes := pem.EncodeToMemory(pemPrivateBlock)
	testKey := string(keyBytes)

	cases := []struct {
		name   string
		args   []string
		env    map[string]string
		expErr string
	}{
		{
			name: "happy_path",
			env: map[string]string{
				"BATCH_SIZE":         "100",
				"EVENTS_PROJECT_ID":  "test-events-project-id",
				"EVENTS_TABLE":       "test-dataset.test-events-table",
				"GITHUB_APP_ID":      "test-github-app-id",
				"GITHUB_INSTALL_ID":  "test-github-install-id",
				"GITHUB_PRIVATE_KEY": testKey,
				"JOB_NAME":           "test-job-name",
				"LEECH_PROJECT_ID":   "test-leech-project-id",
				"LEECH_TABLE":        "test-dataset.test-leech-table",
				"LOGS_BUCKET_NAME":   "test-logs-bucket-name",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, done := context.WithCancel(ctx)
			defer done()

			var cmd LeechCommand
			cmd.testFlagSetOpts = []cli.Option{cli.WithLookupEnv(envconfig.MultiLookuper(
				envconfig.MapLookuper(tc.env),
			).Lookup)}

			_, _, _ = cmd.Pipe()

			_, err := cmd.RunUnstarted(ctx, tc.args, &testObjectWriter{})
			if diff := testutil.DiffErrString(err, tc.expErr); diff != "" {
				t.Fatal(diff)
			}
			if err != nil {
				return
			}
		})
	}
}

type testObjectWriter struct{}

func (w *testObjectWriter) Write(ctx context.Context, reader io.Reader, descriptor string) error {
	return nil
}
