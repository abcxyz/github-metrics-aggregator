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

package retry

import "context"

type retrieveCheckpointIDRes struct {
	res string
	err error
}

type writeCheckpointIDRes struct {
	err error
}

type MockDatastore struct {
	retrieveCheckpointID *retrieveCheckpointIDRes
	writeCheckpointID    *writeCheckpointIDRes
}

func (f *MockDatastore) WriteFailureEvent(ctx context.Context, failureEventTableID, deliveryID, createdAt string) error {
	return nil
}

func (f *MockDatastore) RetrieveCheckpointID(ctx context.Context, checkpointTableID string) (string, error) {
	if f.retrieveCheckpointID != nil {
		return f.retrieveCheckpointID.res, f.retrieveCheckpointID.err
	}
	return "", nil
}

func (f *MockDatastore) WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt string) error {
	if f.writeCheckpointID != nil {
		return f.writeCheckpointID.err
	}
	return nil
}

func (f *MockDatastore) Close() error {
	return nil
}
