// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artifact

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/testutil"
)

func TestPipeline_handleMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name             string
		repoName         string
		bucketName       string
		logPath          string
		gcsPath          string
		wantErr          string
		tokenHandlerFunc http.HandlerFunc
		ghHandlerFunc    http.HandlerFunc
		writerFunc       func(context.Context, io.Reader, string) error
		wantArtifact     string
	}{
		{
			name:         "success",
			repoName:     "test/repo",
			bucketName:   "test",
			logPath:      "test/repo/logs",
			gcsPath:      "gs://test/repo/logs/artifacts.tar.gz",
			wantArtifact: "ok",
		},
		{
			name:       "failed_access_token_generation",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			tokenHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: "failed to get token",
		},
		{
			name:       "github_general_failure",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			ghHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "test bad request")
			},
			wantErr: `error response from GitHub - response body: "test bad request"`,
		},
		{
			name:       "github_logs_not_found",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			ghHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "not found")
			},
			wantErr: "GitHub logs expired",
		},
		{
			name:       "github_logs_gone",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			ghHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusGone)
				fmt.Fprintf(w, "gone")
			},
			wantErr: "GitHub logs expired",
		},
		{
			name:       "object_write_bad_url",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "HOT GARBAGE",
			wantErr:    "error copying logs to cloud storage: malformed gcs url: invalid uri: [HOT GARBAGE]",
		},
		{
			name:       "object_write_failure",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			writerFunc: func(ctx context.Context, r io.Reader, s string) error {
				return fmt.Errorf("write failed")
			},
			wantErr: "error copying logs to cloud storage: write failed",
		},
		{
			name:       "read_write_match",
			repoName:   "test/repo",
			bucketName: "test",
			logPath:    "test/repo/logs",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			ghHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprintf(w, "test-results")
			},
			wantArtifact: "test-results",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.tokenHandlerFunc != nil {
					tc.tokenHandlerFunc(w, r)
					return
				}
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing accept header")
					return
				}
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}
				w.WriteHeader(201)
				fmt.Fprintf(w, `{"token":"this-is-the-token-from-github"}`)
			}))

			testPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatal(err)
			}

			githubApp, err := githubauth.NewApp("test-app-id", "test-install-id", testPrivateKey,
				githubauth.WithAccessTokenURLPattern(fakeTokenServer.URL+"/%s/access_tokens"),
			)
			if err != nil {
				t.Fatal(err)
			}

			writer := testObjectWriter{
				writerFunc: tc.writerFunc,
			}
			ingest := &logIngester{
				bucketName: tc.bucketName,
				githubApp:  githubApp,
				storage:    &writer,
				client:     &http.Client{},
			}

			fakeGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.ghHandlerFunc != nil {
					tc.ghHandlerFunc(w, r)
					return
				}
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing accept header")
					return
				}

				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}

				w.WriteHeader(200)
				fmt.Fprintf(w, "ok")
			}))

			err = ingest.handleMessage(ctx, tc.repoName, fmt.Sprintf("%s/%s", fakeGitHub.URL, tc.logPath), tc.gcsPath)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}

			if got, want := writer.gotArtifact, tc.wantArtifact; got != want {
				t.Errorf("artifacts written got=%v want=%v", got, want)
			}
		})
	}
}

type testObjectWriter struct {
	writerFunc  func(context.Context, io.Reader, string) error
	gotArtifact string
}

func (w *testObjectWriter) Write(ctx context.Context, reader io.Reader, descriptor string) error {
	if w.writerFunc != nil {
		return w.writerFunc(ctx, reader, descriptor)
	}
	if reader == nil {
		return fmt.Errorf("no reader provided")
	}
	if _, _, _, err := parseGCSURI(descriptor); err != nil {
		return fmt.Errorf("malformed gcs url: %w", err)
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}
	w.gotArtifact = string(content)
	return nil
}
