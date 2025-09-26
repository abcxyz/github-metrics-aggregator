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

package retry

import (
	"context"
	"time"
)

type AcquireRes struct {
	Err error
}

type CloseRes struct {
	Err error
}

type MockLock struct {
	AcquireFn func(context.Context, time.Duration) error
	CloseFn   func(context.Context) error
}

func (m *MockLock) Acquire(ctx context.Context, dur time.Duration) error {
	return m.AcquireFn(ctx, dur)
}

func (m *MockLock) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}
