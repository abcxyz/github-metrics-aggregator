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

package main

import (
	"testing"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/pkg/testutil"
)

func TestNewQualifiedTableName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tableName string
		want      *bigqueryio.QualifiedTableName
		wantErr   string
	}{
		{
			name:      "parses_google_sql_format",
			tableName: "project.dataset.table",
			want: &bigqueryio.QualifiedTableName{
				Project: "project",
				Dataset: "dataset",
				Table:   "table",
			},
		},
		{
			name:      "rejects_legacy_sql_format",
			tableName: "project:dataset.table",
			wantErr:   "expected 3 parts separated by a . but got 2",
		},
		{
			name:      "rejects_missing_first_period",
			tableName: "projectdataset.table",
			wantErr:   "expected 3 parts separated by a . but got 2",
		},
		{
			name:      "rejects_missing_second_period",
			tableName: "project.datasettable",
			wantErr:   "expected 3 parts separated by a . but got 2",
		},
		{
			name:      "rejects_missing_both_periods",
			tableName: "projectdatasettable",
			wantErr:   "expected 3 parts separated by a . but got 1",
		},
		{
			name:      "rejects_empty_string",
			tableName: "",
			wantErr:   "expected 3 parts separated by a . but got 1",
		},
		{
			name:      "rejects_missing_project",
			tableName: ".dataset.table",
			wantErr:   "invalid project id ``",
		},
		{
			name:      "rejects_missing_dataset",
			tableName: "project..table",
			wantErr:   "invalid dataset id ``",
		},
		{
			name:      "rejects_missing_table",
			tableName: "project.dataset.",
			wantErr:   "invalid table name ``",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := newQualifiedTableName(tc.tableName)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("newQualifiedTableName got unexpected result (-got,+want):\n%s", diff)
			}
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("newQualifiedTableName got unexpected error (-got,+want):\n%s", diff)
			}
		})
	}
}
