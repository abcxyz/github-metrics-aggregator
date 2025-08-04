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

package kms

import (
	"context"
	"fmt"

	"google.golang.org/api/option"

	"github.com/sethvargo/go-gcpkms/pkg/gcpkms"

	kms "cloud.google.com/go/kms/apiv1"
)

// KeyManagement provides a client and dataset identifiers.
type KeyManagement struct {
	client *kms.KeyManagementClient
}

// NewKeyManagementClient creates a new instance of a KMS client.
func NewKeyManagement(ctx context.Context, opts ...option.ClientOption) (*KeyManagement, error) {
	client, err := kms.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new key management client: %w", err)
	}

	return &KeyManagement{
		client: client,
	}, nil
}

// CreateSigner leverages the gcpkms package to create a signer.
func (km *KeyManagement) CreateSigner(ctx context.Context, kmsAppPrivateKeyID string) (*gcpkms.Signer, error) {
	signer, err := gcpkms.NewSigner(ctx, km.client, kmsAppPrivateKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to create app signer: %w", err)
	}
	return signer, nil
}

// Close releases any resources held by the KeyManagement client.
func (km *KeyManagement) Close() error {
	if err := km.client.Close(); err != nil {
		return fmt.Errorf("failed to close KeyManagement client: %w", err)
	}
	return nil
}
