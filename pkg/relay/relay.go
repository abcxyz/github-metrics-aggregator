// Copyright 2025 The Authors (see AUTHORS file)
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

// Package relay contains the HTTP handler and logic for the relay service.
package relay

import (
	"io"
	"net/http"

	"github.com/abcxyz/pkg/logging"
)

const mb = 1 << 20

type wrappedMessage struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
		ID         string            `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func (s *Server) handleRelay() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.FromContext(ctx)

		defer r.Body.Close()

		req, err := io.ReadAll(io.LimitReader(r.Body, 25*mb))
		if err != nil {
			logger.ErrorContext(ctx, "failed to read request body", "error", err)
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		data, attrs, err := s.messageEnricher.Enrich(ctx, req)
		if err != nil {
			logger.ErrorContext(ctx, "failed to enrich event", "error", err)
			http.Error(w, "failed to enrich event", http.StatusInternalServerError)
			return
		}

		if err := s.relayMessenger.Send(ctx, data, attrs); err != nil {
			logger.ErrorContext(ctx, "failed to write message to relay topic",
				"code", http.StatusInternalServerError,
				"body", "error writing message to relay topic",
				"error", err)
		}

		w.WriteHeader(http.StatusOK)
	})
}
