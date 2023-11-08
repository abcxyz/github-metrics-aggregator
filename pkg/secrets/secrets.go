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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// the block type for a PEM encoded rsa private key.
const pemRSAPrivateKey = "RSA PRIVATE KEY"

// AccessSecretFromSecretManager reads a secret from Secret Manager and validates that it was not
// corrupted during retrieval. The secretResourceName should be in the format:
// 'projects/*/secrets/*/versions/*'. This function is intended for use cases
// where you need to fetch one and only one secret from secret manager as it
// instantiates a temporary secret manager client in order to fetch the secret.
// Due to the expensive nature of instantiating clients, AccessSecret should be
// used instead if multiple secrets need to be fetched from secret manager.
func AccessSecretFromSecretManager(ctx context.Context, secretResourceName string) (s string, e error) {
	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer func(sm *secretmanager.Client) {
		err := sm.Close()
		if err != nil {
			e = fmt.Errorf("failed to close secret manager client: %w", err)
		}
	}(sm)
	secret, err := AccessSecret(ctx, sm, secretResourceName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
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

// ParsePrivateKey parses the PEM encoded RSA private key into a rsa.PrivateKey.
func ParsePrivateKey(privateKeyContent string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyContent))
	if block == nil || block.Type != pemRSAPrivateKey {
		return nil, fmt.Errorf("failed to decode PEM RSA private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
	}
	return key, nil
}
