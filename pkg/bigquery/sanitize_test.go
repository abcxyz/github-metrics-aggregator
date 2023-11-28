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

package bigquery

import (
	"strings"
	"testing"
)

func genStringLen(num int, rune rune) string {
	if num < 0 {
		panic("cannot make negative length string")
	}
	var sb strings.Builder
	sb.Grow(num)
	for i := 0; i < num; i++ {
		sb.WriteRune(rune)
	}
	return sb.String()
}

func Test_validateGCPProjectID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "happy_path",
			input:   "github4-metrics",
			wantErr: false,
		},
		{
			name:    "space_fails",
			input:   " ",
			wantErr: true,
		},
		{
			name:    "end_with_hyphen_fails",
			input:   "foobar-",
			wantErr: true,
		},
		{
			name:    "start_with_hyphen_fails",
			input:   "-foobar",
			wantErr: true,
		},
		{
			name:    "start_with_number_fails",
			input:   "4foobar",
			wantErr: true,
		},
		{
			name:    "too_short_fails",
			input:   "short",
			wantErr: true,
		},
		{
			name:    "six_chars_succeeds",
			input:   "abcdef",
			wantErr: false,
		},
		{
			name:    "end_with_number_succeeds",
			input:   "foobar4",
			wantErr: false,
		},
		{
			name:    "empty_string_fails",
			input:   "",
			wantErr: true,
		},
		{
			name:    "underscore_fails",
			input:   "github_metrics",
			wantErr: true,
		},
		{
			name:    "closing_backtick_fails",
			input:   "github_metrics`",
			wantErr: true,
		},
		{
			name:    "closing_quote_fails",
			input:   "github_metrics\"",
			wantErr: true,
		},
		{
			name:    "semi_colon_fails",
			input:   "github_metrics;",
			wantErr: true,
		},
		{
			name:    "space_fails",
			input:   "github_metrics ",
			wantErr: true,
		},
		{
			name:    "dot_fails",
			input:   "github.metrics",
			wantErr: true,
		},
		{
			name:    "unicode_fails",
			input:   "様様様様様様",
			wantErr: true,
		},
		{
			name:    "empty_fails",
			input:   "",
			wantErr: true,
		},
		{
			name:    "30_succeeds",
			input:   genStringLen(30, 'a'),
			wantErr: false,
		},
		{
			name:    "too_long_fails",
			input:   genStringLen(31, 'a'),
			wantErr: true,
		},
		{
			name:    "multiline_string_fails",
			input:   "foobar\nfoobar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateGCPProjectID(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validateGCPProjectID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateDatasetID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "happy_path",
			input:   "github_metrics",
			wantErr: false,
		},
		{
			name:    "space_fails",
			input:   " ",
			wantErr: true,
		},
		{
			name:    "empty_string_fails",
			input:   "",
			wantErr: true,
		},
		{
			name:    "hyphen_fails",
			input:   "github-metrics",
			wantErr: true,
		},
		{
			name:    "closing_backtick_fails",
			input:   "github_metrics`",
			wantErr: true,
		},
		{
			name:    "closing_quote_fails",
			input:   "github_metrics\"",
			wantErr: true,
		},
		{
			name:    "semi_colon_fails",
			input:   "github_metrics;",
			wantErr: true,
		},
		{
			name:    "space_fails",
			input:   "github_metrics ",
			wantErr: true,
		},
		{
			name:    "dot_fails",
			input:   "github.metrics",
			wantErr: true,
		},
		{
			name:    "unicode_fails",
			input:   "様",
			wantErr: true,
		},
		{
			name:    "empty_fails",
			input:   "",
			wantErr: true,
		},
		{
			name:    "single_char_succeeds",
			input:   "_",
			wantErr: false,
		},
		{
			name:    "1024_succeeds",
			input:   genStringLen(1024, 'a'),
			wantErr: false,
		},
		{
			name:    "too_long_fails",
			input:   genStringLen(1025, 'a'),
			wantErr: true,
		},
		{
			name:    "multiline_string_fails",
			input:   "foobar\nfoobar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateDatasetID(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validateDatasetID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateTableName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "space_passes",
			input:   "github_metrics ",
			wantErr: false,
		},
		{
			name:    "only_space_passes",
			input:   "         ",
			wantErr: false,
		},
		{
			name:    "empty_string_fails",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unicode_categories_pass",
			input:   "letter_ʱڰƻǈΞℯ_mark_ॊ⃣ٖ_number_꧓Ⅻ൝5_Pc__‿_Pd_-〰𐺭_Zs_   　",
			wantErr: false,
		},
		{
			name:    "just_right_succeeds_unicode",
			input:   genStringLen(341, '様') + "f",
			wantErr: false,
		},
		{
			name:    "too_long_fails_unicode",
			input:   genStringLen(342, '様'),
			wantErr: true,
		},
		{
			name:    "closing_backtick_fails",
			input:   "github_metrics`",
			wantErr: true,
		},
		{
			name:    "closing_quote_fails",
			input:   "github_metrics\"",
			wantErr: true,
		},
		{
			name:    "semi_colon_fails",
			input:   "github_metrics;",
			wantErr: true,
		},
		{
			name:    "dot_fails",
			input:   "github.metrics",
			wantErr: true,
		},
		{
			name:    "multiline_string_fails",
			input:   "foobar\nfoobar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateTableName(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validateTableName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
