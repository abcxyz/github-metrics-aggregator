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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/pointer"
	"github.com/abcxyz/pkg/testutil"
)

func TestPipeline_handleMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name         string
		bucketName   string
		gcsPath      string
		wantErr      string
		tokenHandler http.HandlerFunc
		logsHandler  http.HandlerFunc
		writerFunc   func(context.Context, io.Reader, string) (string, error)
		wantArtifact string
	}{
		{
			name:         "success",
			bucketName:   "test",
			gcsPath:      "gs://test/repo/logs/artifacts.tar.gz",
			wantArtifact: "ok",
		},
		{
			name:       "failed_access_token_generation",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			tokenHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: "failed to get github access token",
		},
		{
			name:       "github_general_failure",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			logsHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "test bad request")
			},
			wantErr: `error response from GitHub - response body: "test bad request"`,
		},
		{
			name:       "github_logs_not_found",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			logsHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "not found")
			},
			wantErr: "GitHub logs expired",
		},
		{
			name:       "github_logs_gone",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			logsHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusGone)
				fmt.Fprintf(w, "gone")
			},
			wantErr: "GitHub logs expired",
		},
		{
			name:       "object_write_bad_url",
			bucketName: "test",
			gcsPath:    "HOT GARBAGE",
			wantErr:    "error copying logs to cloud storage: malformed gcs url: invalid uri: [HOT GARBAGE]",
		},
		{
			name:       "object_write_failure",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			writerFunc: func(ctx context.Context, r io.Reader, s string) (string, error) {
				return "", fmt.Errorf("write failed")
			},
			wantErr: "error copying logs to cloud storage: write failed",
		},
		{
			name:       "read_write_match",
			bucketName: "test",
			gcsPath:    "gs://test/repo/logs/artifacts.tar.gz",
			logsHandler: func(w http.ResponseWriter, r *http.Request) {
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

			fakeGitHub := func() *httptest.Server {
				mux := http.NewServeMux()
				mux.Handle("GET /app/installations/123", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{"access_tokens_url": "http://%s/app/installations/123/access_tokens"}`, r.Host)
				}))
				mux.Handle("POST /app/installations/123/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.tokenHandler != nil {
						tc.tokenHandler(w, r)
						return
					}
					w.WriteHeader(201)
					fmt.Fprintf(w, `{"token": "this-is-the-token-from-github"}`)
				}))
				mux.Handle("GET /test/repo/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
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

					if tc.logsHandler != nil {
						tc.logsHandler(w, r)
						return
					}
					fmt.Fprintf(w, "ok")
				}))

				return httptest.NewServer(mux)
			}()
			t.Cleanup(func() {
				fakeGitHub.Close()
			})

			testPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatal(err)
			}

			privateKeyPem := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(testPrivateKey),
			})
			ghClient, err := githubclient.NewWithPermissions(ctx, "test-app-id", "123", string(privateKeyPem), map[string]string{
				"actions": "read",
			}, githubauth.WithBaseURL(fakeGitHub.URL))
			if err != nil {
				t.Fatal(err)
			}

			ghClient, err = ghClient.WithBaseURL(fakeGitHub.URL)
			if err != nil {
				t.Fatal(err)
			}

			writer := testObjectWriter{
				writerFunc: tc.writerFunc,
			}

			ingest := logIngester{
				bucketName: tc.bucketName,
				storage:    &writer,
				ghClient:   ghClient,
			}

			_, err = ingest.handleMessage(ctx, fmt.Sprintf("%s/%s", fakeGitHub.URL, "test/repo/logs"), tc.gcsPath)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Process(%+v) got unexpected err: %s", tc.name, diff)
			}

			if got, want := writer.gotArtifact, tc.wantArtifact; got != want {
				t.Errorf("artifacts written got=%v want=%v", got, want)
			}
		})
	}
}

func TestPipeline_commentArtifactOnPRs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name                  string
		bucketName            string
		event                 EventRecord
		artifactStatus        string
		tokenHandler          http.HandlerFunc
		commentResponseStatus *int
		wantErr               string
		expectedCommentCount  int
	}{
		{
			name:       "success",
			bucketName: "test",
			event: EventRecord{
				DeliveryID:         "123",
				RepositorySlug:     "testorg/testrepo",
				RepositoryName:     "testrepo",
				OrganizationName:   "testorg",
				LogsURL:            "https://api.github.com/repos/testorg/testrepo/actions/runs/987/logs",
				GitHubActor:        "user",
				WorkflowURL:        "https://api.github.com/repos/testorg/testrepo/actions/runs/987",
				WorkflowRunID:      "987",
				WorkflowRunAttempt: "1",
				PullRequestNumbers: []string{"456"},
			},
			artifactStatus:       "SUCCESS",
			expectedCommentCount: 1,
		},
		{
			name:       "skip-on-bad-artifact-status",
			bucketName: "test",
			event: EventRecord{
				DeliveryID:         "123",
				RepositorySlug:     "testorg/testrepo",
				RepositoryName:     "testrepo",
				OrganizationName:   "testorg",
				LogsURL:            "https://api.github.com/repos/testorg/testrepo/actions/runs/987/logs",
				GitHubActor:        "user",
				WorkflowURL:        "https://api.github.com/repos/testorg/testrepo/actions/runs/987",
				WorkflowRunID:      "987",
				WorkflowRunAttempt: "1",
				PullRequestNumbers: []string{"456"},
			},
			artifactStatus:       "FAILURE",
			expectedCommentCount: 0,
		},
		{
			name:       "fail-bad-pr-number",
			bucketName: "test",
			event: EventRecord{
				DeliveryID:         "123",
				RepositorySlug:     "testorg/testrepo",
				RepositoryName:     "testrepo",
				OrganizationName:   "testorg",
				LogsURL:            "https://api.github.com/repos/testorg/testrepo/actions/runs/987/logs",
				GitHubActor:        "user",
				WorkflowURL:        "https://api.github.com/repos/testorg/testrepo/actions/runs/987",
				WorkflowRunID:      "987",
				WorkflowRunAttempt: "1",
				PullRequestNumbers: []string{"456blahblahblah"},
			},
			artifactStatus:       "SUCCESS",
			expectedCommentCount: 0,
			wantErr:              "error parsing pr number from event payload",
		},
		{
			name:       "fail-unexpected-response",
			bucketName: "test",
			event: EventRecord{
				DeliveryID:         "123",
				RepositorySlug:     "testorg/testrepo",
				RepositoryName:     "testrepo",
				OrganizationName:   "testorg",
				LogsURL:            "https://api.github.com/repos/testorg/testrepo/actions/runs/987/logs",
				GitHubActor:        "user",
				WorkflowURL:        "https://api.github.com/repos/testorg/testrepo/actions/runs/987",
				WorkflowRunID:      "987",
				WorkflowRunAttempt: "1",
				PullRequestNumbers: []string{"456"},
			},
			artifactStatus:        "SUCCESS",
			commentResponseStatus: pointer.To(401),
			expectedCommentCount:  1,
			wantErr:               "error commenting artifact on pull request",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			commentRequestCount := 0
			fakeGitHub := func() *httptest.Server {
				mux := http.NewServeMux()
				mux.Handle("GET /app/installations/123", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{"access_tokens_url": "http://%s/app/installations/123/access_tokens"}`, r.Host)
				}))
				mux.Handle("POST /app/installations/123/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.tokenHandler != nil {
						tc.tokenHandler(w, r)
						return
					}
					w.WriteHeader(201)
					fmt.Fprintf(w, `{"token": "this-is-the-token-from-github"}`)
				}))
				mux.Handle("POST /api/v3/repos/testorg/testrepo/issues/456/comments", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					commentRequestCount += 1
					if tc.commentResponseStatus != nil {
						w.WriteHeader(*tc.commentResponseStatus)
					} else {
						w.WriteHeader(201)
					}
				}))

				return httptest.NewServer(mux)
			}()
			t.Cleanup(func() {
				fakeGitHub.Close()
			})

			testPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatal(err)
			}

			privateKeyPem := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(testPrivateKey),
			})
			ghClient, err := githubclient.NewWithPermissions(ctx, "test-app-id", "123", string(privateKeyPem), map[string]string{
				"pull_requests": "write",
			}, githubauth.WithBaseURL(fakeGitHub.URL))
			if err != nil {
				t.Fatal(err)
			}

			ghClient, err = ghClient.WithBaseURL(fakeGitHub.URL)
			if err != nil {
				t.Fatal(err)
			}

			ingest := logIngester{
				bucketName: tc.bucketName,
				ghClient:   ghClient,
			}

			artifact := ArtifactRecord{
				DeliveryID:       tc.event.DeliveryID,
				ProcessedAt:      time.Now(),
				Status:           tc.artifactStatus,
				WorkflowURI:      tc.event.WorkflowURL,
				LogsURI:          fmt.Sprintf("gs://%s/%s/%s/artifacts.tar.gz", tc.bucketName, tc.event.RepositorySlug, tc.event.DeliveryID),
				GitHubActor:      tc.event.GitHubActor,
				OrganizationName: tc.event.OrganizationName,
				RepositoryName:   tc.event.RepositoryName,
				RepositorySlug:   tc.event.RepositorySlug,
				JobName:          "testjob",
			}

			err = ingest.commentArtifactOnPRs(ctx, &tc.event, &artifact, "testurl")
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("commentArtifactOnPRs(%+v) got unexpected err: %s", tc.name, diff)
			}
			if tc.expectedCommentCount != commentRequestCount {
				t.Errorf("commentArtifactOnPRs(%+v) expected to make %d CommentPR API calls but instead made %d", tc.name, tc.expectedCommentCount, commentRequestCount)
			}
		})
	}
}

type testObjectWriter struct {
	writerFunc  func(context.Context, io.Reader, string) (string, error)
	gotArtifact string
}

func (w *testObjectWriter) Write(ctx context.Context, reader io.Reader, descriptor string) (string, error) {
	if w.writerFunc != nil {
		return w.writerFunc(ctx, reader, descriptor)
	}
	if reader == nil {
		return "", fmt.Errorf("no reader provided")
	}
	if _, _, _, err := parseGCSURI(descriptor); err != nil {
		return "", fmt.Errorf("malformed gcs url: %w", err)
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	w.gotArtifact = string(content)
	return "testartifacturl", nil
}
