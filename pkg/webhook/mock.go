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

package webhook

import "context"

type deliveryEventExistsRes struct {
	res bool
	err error
}

type failureEventsExceedsRetryLimitRes struct {
	res bool
	err error
}

type MockDatastore struct {
	deliveryEventExists            *deliveryEventExistsRes
	failureEventsExceedsRetryLimit *failureEventsExceedsRetryLimitRes
}

func (f *MockDatastore) DeliveryEventExists(ctx context.Context, eventsTableID, deliveryID string) (bool, error) {
	if f.deliveryEventExists != nil {
		return f.deliveryEventExists.res, f.deliveryEventExists.err
	}
	return false, nil
}

func (f *MockDatastore) FailureEventsExceedsRetryLimit(ctx context.Context, failureEventTableID, deliveryID string, retryLimit int) (bool, error) {
	if f.failureEventsExceedsRetryLimit != nil {
		return f.failureEventsExceedsRetryLimit.res, f.failureEventsExceedsRetryLimit.err
	}
	return false, nil
}

func (f *MockDatastore) WriteFailureEvent(ctx context.Context, failureEventTableID, deliveryID, createdAt string) error {
	return nil
}

func (f *MockDatastore) RetrieveCheckpointID(ctx context.Context, checkpointTableID string) (string, error) {
	return "test-checkpoint-id", nil
}

func (f *MockDatastore) WriteCheckpointID(ctx context.Context, checkpointTableID, deliveryID, createdAt string) error {
	return nil
}

func (f *MockDatastore) Shutdown() error {
	return nil
}
