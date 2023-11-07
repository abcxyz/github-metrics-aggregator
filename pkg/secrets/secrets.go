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

package secrets

import (
	"context"
	"crypto/rsa"
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// GetSecret reads a secret from Secret Manager and validates that it was not
// corrupted during retrieval. The secretResourceName should be in the format:
// 'projects/*/secrets/*/versions/*'.
func GetSecret(ctx context.Context, secretResourceName string) (string, error) {
	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %w", err)
	}
	secret, err := AccessSecret(ctx, sm, secretResourceName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}
	if err := sm.Close(); err != nil {
		return "", fmt.Errorf("failed to close secret manager client: %w", err)
	}
	return secret, nil
}

// AccessSecret reads a secret from Secret Manager using the given client and
// validates that it was not corrupted during retrieval. The secretResourceName
// should be in the format: 'projects/*/secrets/*/versions/*'.
func AccessSecret(ctx context.Context, client *secretmanager.Client, secretResourceName string) (string, error) {
	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: secretResourceName,
	}
	result, err := client.AccessSecretVersion(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("failed to get privateKeyResourceName version for %q - %w", secretResourceName, err)
	}
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("failed to get privateKeyResourceName version for %q - data corrupted", secretResourceName)
	}
	return string(result.Payload.Data), nil
}

// ParsePrivateKey parses a string containing a PEM encoded private key into
// a rsa.PrivateKey struct.
func ParsePrivateKey(privateKeyContent string) (*rsa.PrivateKey, error) {
	parsedKey, _, err := jwk.DecodePEM([]byte(privateKeyContent))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
	}
	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
	}
	return privateKey, nil
}
