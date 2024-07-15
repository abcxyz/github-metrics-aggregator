// Copyright 2024 The Authors (see AUTHORS file)
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

package bq

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
)

type TestStruct struct {
	StringField string
	IntField    int
}

type TestRowIterator struct {
	rows []TestStruct
}

func (ri *TestRowIterator) Next(i interface{}) error {
	if len(ri.rows) == 0 {
		return iterator.Done
	}
	t, ok := i.(*TestStruct)
	if !ok {
		return fmt.Errorf("invalid type %T", i)
	}
	t.StringField = ri.rows[0].StringField
	t.IntField = ri.rows[0].IntField
	ri.rows = ri.rows[1:]
	return nil
}

func TestRowsToSlice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		rows      rowItr[TestStruct]
		totalRows uint64
		want      []*TestStruct
	}{
		{
			name: "no_items",
			rows: &TestRowIterator{
				rows: []TestStruct{},
			},
			totalRows: 0,
			want:      []*TestStruct{},
		},
		{
			name: "one_item",
			rows: &TestRowIterator{
				rows: []TestStruct{
					{
						StringField: "a",
						IntField:    1,
					},
				},
			},
			totalRows: 1,
			want: []*TestStruct{
				{
					StringField: "a",
					IntField:    1,
				},
			},
		},
		{
			name: "several_items",
			rows: &TestRowIterator{
				rows: []TestStruct{
					{
						StringField: "a",
						IntField:    1,
					},
					{
						StringField: "b",
						IntField:    2,
					},
					{
						StringField: "c",
						IntField:    3,
					},
				},
			},
			totalRows: 2,
			want: []*TestStruct{
				{
					StringField: "a",
					IntField:    1,
				},
				{
					StringField: "b",
					IntField:    2,
				},
				{
					StringField: "c",
					IntField:    3,
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, _ := rowsToSlice[TestStruct](tc.rows, tc.totalRows)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("unexpected result (-got, +want):\n%s", diff)
			}
		})
	}
}
