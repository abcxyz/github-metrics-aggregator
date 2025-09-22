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

type failureEventsExceedsRetryLimitRes struct {
	res bool
	err error
}

type MockDatastore struct {
	failureEventsExceedsRetryLimit *failureEventsExceedsRetryLimitRes
}

func (m *MockDatastore) FailureEventsExceedsRetryLimit(ctx context.Context, failureEventTableID, deliveryID string, retryLimit int) (bool, error) {
	if m.failureEventsExceedsRetryLimit != nil {
		return m.failureEventsExceedsRetryLimit.res, m.failureEventsExceedsRetryLimit.err
	}
	return false, nil
}

func (m *MockDatastore) WriteFailureEvent(ctx context.Context, failureEventTableID, deliveryID, createdAt string) error {
	return nil
}

func (m *MockDatastore) Close() error {
	return nil
}
