// Copyright 2023 Google LLC
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

//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative pubsub_schemas/event.proto
//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative pubsub_schemas/enriched_event.proto

// Package protos contains versioned protos for the github-metrics-aggregator.
package protos
