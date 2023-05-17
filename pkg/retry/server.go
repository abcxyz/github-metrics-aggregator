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

// Package retry is the retry server responsible for interacting with GitHub
// APIs to retry failed events.
package retry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/abcxyz/pkg/healthcheck"
	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-github/v52/github"
	"github.com/sethvargo/go-gcslock"
	"google.golang.org/api/option"
)

// Datastore adheres to the interaction the retry service has with a datastore.
type Datastore interface {
	RetrieveCheckpointID(ctx context.Context, checkpointTableID string) (string, error)
	WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt string) error
	DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error)
	Close() error
}

// GitHubSource aheres to the interaction the retyr service has with GitHub APIs.
type GitHubSource interface {
	ListDeliveries(ctx context.Context, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error)
	RedeliverEvent(ctx context.Context, deliveryID int64) error
}

type Server struct {
	datastore         Datastore
	gcsLock           gcslock.Lockable
	github            GitHubSource
	lockTTL           time.Duration
	checkpointTableID string
	eventsTableID     string
	projectID         string
}

// RetryClientOptions encapsulate client config options as well as dependency
// implementation overrides.
type RetryClientOptions struct {
	BigQueryClientOpts      []option.ClientOption
	GCSLockClientOpts       []option.ClientOption
	DatastoreClientOverride Datastore        // used for unit testing
	GCSLockClientOverride   gcslock.Lockable // used for unit testing
	GitHubOverride          GitHubSource     // used for unit testing
}

// NewServer creates a new HTTP server implementation that will handle
// communication with GitHub APIs.
func NewServer(ctx context.Context, cfg *Config, rco *RetryClientOptions) (*Server, error) {
	datastore := rco.DatastoreClientOverride
	if datastore == nil {
		bq, err := NewBigQuery(ctx, cfg.BigQueryProjectID, cfg.DatasetID, rco.BigQueryClientOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize BigQuery client: %w", err)
		}
		datastore = bq
	}

	gcsLock := rco.GCSLockClientOverride
	if gcsLock == nil {
		lock, err := gcslock.New(ctx, cfg.BucketName, "retry-lock", rco.GCSLockClientOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain GCS lock: %w", err)
		}
		gcsLock = lock
	}

	github := rco.GitHubOverride
	if github == nil {
		gh, err := NewGitHub(ctx, cfg.GitHubAppID, cfg.GitHubPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize github client: %w", err)
		}
		github = gh
	}

	return &Server{
		datastore:         datastore,
		gcsLock:           gcsLock,
		github:            github,
		projectID:         cfg.ProjectID,
		lockTTL:           cfg.LockTTL,
		checkpointTableID: cfg.CheckpointTableID,
		eventsTableID:     cfg.EventsTableID,
	}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes(ctx context.Context) http.Handler {
	logger := logging.FromContext(ctx)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthcheck.HandleHTTPHealthCheck())
	mux.Handle("/retry", s.handleRetry())

	// Middleware
	root := logging.HTTPInterceptor(logger, s.projectID)(mux)

	return root
}

// Close handles the graceful shutdown of the retry server.
func (s *Server) Close() error {
	if err := s.datastore.Close(); err != nil {
		return fmt.Errorf("failed to shutdown the BigQuery connection: %w", err)
	}

	if err := s.gcsLock.Close(context.Background()); err != nil {
		return fmt.Errorf("failed to close the GCS lock connection: %w", err)
	}

	return nil
}
