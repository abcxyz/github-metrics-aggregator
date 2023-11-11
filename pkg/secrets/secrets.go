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
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// AccessSecret reads a secret from Secret Manager using the given client and
// validates that it was not corrupted during retrieval. The secretResourceName
// should be in the format: 'projects/*/secrets/*/versions/*'.
func AccessSecret(ctx context.Context, client *secretmanager.Client, secretResourceName string) (string, error) {
	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretResourceName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretResourceName, err)
	}
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("failed to access secret %s: data corrupted", secretResourceName)
	}
	return string(result.Payload.Data), nil
}
