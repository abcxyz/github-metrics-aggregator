// Copyright 2026 The Authors (see AUTHORS file)
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

package events

type EnrichedEvent struct {
	DeliveryId       string `json:"delivery_id,omitempty"`
	Signature        string `json:"signature,omitempty"`
	Received         string `json:"received,omitempty"`
	Event            string `json:"event,omitempty"`
	Payload          string `json:"payload,omitempty"`
	EnterpriseId     string `json:"enterprise_id,omitempty"`
	EnterpriseName   string `json:"enterprise_name,omitempty"`
	OrganizationId   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	RepositoryId     string `json:"repository_id,omitempty"`
	RepositoryName   string `json:"repository_name,omitempty"`
}
