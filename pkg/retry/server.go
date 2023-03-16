// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package retry is the retry server responsible for interacting with GitHub APIs to retry failed events
package retry

import (
	"context"
	"fmt"
	"net/http"
)

// TODO provide more fields to this struct.
type Server struct{}

// NewServer creates a new HTTP server implementation that will handle
// communication with GitHub APIs.
func NewServer(ctx context.Context, cfg *Config) (*Server, error) {
	return &Server{}, nil
}

// Routes creates a ServeMux of all of the routes that
// this Router supports.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/retry", s.handleRetry())
	return mux
}

// handleRetry handles calling GitHub APIs to search and retry for failed events.
func (s *Server) handleRetry() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `TODO: Make GitHub API calls and other stuff.. \n`)
	})
}
