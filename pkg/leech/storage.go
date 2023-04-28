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

package leech

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
)

type ObjectStore struct {
	client *storage.Client
}

// NewObjectClient creates a new cloud storage client.
func NewObjectStore(ctx context.Context) (*ObjectStore, error) {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializaing cloud storage client: %w", err)
	}
	return &ObjectStore{client: sc}, nil
}

// WriteObject writes an object to Google Cloud Storage.
func (s *ObjectStore) WriteObject(ctx context.Context, content io.Reader, objectDescriptor string) error {
	// Split the descriptor into chunks
	bucketName, objectName, _, err := parseGCSURI(objectDescriptor)
	if err != nil {
		return fmt.Errorf("failed to parse gcs uri: %w", err)
	}

	// Connect to bucket
	bucket := s.client.Bucket(bucketName)
	// Setup the GCS object with the filename to write to
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		return fmt.Errorf("failed to copy contents of reader to cloud storage object: %w", err)
	}

	// File appears in GCS after Close
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close gcs file: %w", err)
	}

	return nil
}

// parseGCSURI parses a gcs uri of the type gs://blah/blah/blah.blah
// The parts are:
//
//	bucket name
//	object path
//	file name
//
// Throws an error if the uri cannot be parsed.
func parseGCSURI(objectURI string) (string, string, string, error) {
	// First verify that all of the parts exist
	r, _ := regexp.Compile("gs://(.*)/(.*)")
	if !r.MatchString(objectURI) {
		return "", "", "", fmt.Errorf("invalid uri: [%s]", objectURI)
	}
	// Extract bucket name by splitting string by '/'
	// take the 3rd item in the list (index position 2) which is the bucket name
	parts := strings.Split(objectURI, "/")
	bucket := parts[2]

	// Extract object name by splitting string to remove gs:// prefix and bucket name
	// rejoin to rebuild the file path
	objectName := strings.Join(parts[3:], "/")
	// Extract the last segment as the filename
	fileName := parts[len(parts)-1]
	return bucket, objectName, fileName, nil
}
