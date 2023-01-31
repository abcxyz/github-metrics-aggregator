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

package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type config struct {
	ProjectID              string        `env:"PROJECT_ID,required"`
	DatasetID              string        `env:"DATASET_ID,required"`
	IDToken                string        `env:"ID_TOKEN,required"`
	GitHubWebhookSecret    string        `env:"GITHUB_WEBHOOK_SECRET,required"`
	EndpointURL            string        `env:"ENDPOINT_URL,required"`
	HTTPRequestTimeout     time.Duration `env:"HTTP_REQUEST_TIMEOUT,default=60s"`
	QueryRetryWaitDuration time.Duration `env:"QUERY_RETRY_WAIT_DURATION,default=5s"`
	QueryRetryLimit        uint64        `env:"QUERY_RETRY_COUNT,default=5"`
}

func newTestConfig(ctx context.Context) (*config, error) {
	var c config
	if err := envconfig.ProcessWith(ctx, &c, envconfig.OsLookuper()); err != nil {
		return nil, fmt.Errorf("failed to process environment: %w", err)
	}
	return &c, nil
}
