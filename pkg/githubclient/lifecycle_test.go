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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Lifecycle(t *testing.T) {
	t.Parallel()

	// 1. Generate a dummy private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	// 2. Setup mock GitHub Enterprise server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			// We can't really fail the test from here cleanly without the t context,
			// but we can log or panic if strictly needed, or just satisfy the linter.
			// Ideally we log it.
			fmt.Printf("failed to write response: %v\n", err)
		}
	}))
	defer ts.Close()

	// 3. Create context that we will cancel
	ctx, cancel := context.WithCancel(context.Background())

	// 4. Initialize Client
	cfg := &Config{
		GitHubAppID:               "123",
		GitHubPrivateKey:          string(pemBytes),
		GitHubEnterpriseServerURL: ts.URL, // Point to our mock server
	}

	client, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 5. Cancel the context immediately
	cancel()

	// 6. Verify client still works
	// We use ListDeliveries as it triggers a request using the underlying client
	// Note: We need a fresh context for the REQUEST itself, but the CLIENT shouldn't be dead.
	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	// The client internal transport should NOT be bound to the dead 'ctx'.
	// If it was bound, this request would fail immediately or fails to dial.
	_, _, err = client.ListDeliveries(reqCtx, nil)
	if err != nil {
		t.Fatalf("Client failed to make request after init context cancellation: %v", err)
	}
}
