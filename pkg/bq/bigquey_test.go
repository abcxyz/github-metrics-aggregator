package bq

import (
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
	"testing"
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
	var t *TestStruct = i.(*TestStruct)
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
