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

import (
	"context"

	"github.com/google/go-github/v56/github"
)

type listDeliveriesRes struct {
	deliveries []*github.HookDelivery
	res        *github.Response
	err        error
}

type redeliverEventRes struct {
	err error
}

type MockGitHub struct {
	listDeliveries *listDeliveriesRes
	redeliverEvent *redeliverEventRes
}

func (m *MockGitHub) ListDeliveries(ctx context.Context, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error) {
	if m.listDeliveries != nil {
		return m.listDeliveries.deliveries, m.listDeliveries.res, m.listDeliveries.err
	}
	return []*github.HookDelivery{}, nil, nil
}

func (m *MockGitHub) RedeliverEvent(ctx context.Context, deliveryID int64) error {
	if m.redeliverEvent != nil {
		return m.redeliverEvent.err
	}
	return nil
}
