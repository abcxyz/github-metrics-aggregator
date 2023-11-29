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

func TestValidateGCPProjectID(t *testing.T) {
	t.Parallel()
	cases := []struct {
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
			input:   strings.Repeat("a", 30),
			wantErr: false,
		},
		{
			name:    "too_long_fails",
			input:   strings.Repeat("a", 31),
			wantErr: true,
		},
		{
			name:    "multiline_string_fails",
			input:   "foobar\nbarfoo",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateGCPProjectID(tc.input); (err != nil) != tc.wantErr {
				t.Errorf("ValidateGCPProjectID() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateDatasetID(t *testing.T) {
	t.Parallel()
	cases := []struct {
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
			input:   strings.Repeat("a", 1024),
			wantErr: false,
		},
		{
			name:    "too_long_fails",
			input:   strings.Repeat("a", 1025),
			wantErr: true,
		},
		{
			name:    "multiline_string_fails",
			input:   "foobar\nbarfoo",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateDatasetID(tc.input); (err != nil) != tc.wantErr {
				t.Errorf("ValidateDatasetID() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateTableName(t *testing.T) {
	t.Parallel()
	cases := []struct {
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
			input:   strings.Repeat("様", 341) + "f",
			wantErr: false,
		},
		{
			name:    "too_long_fails_unicode",
			input:   strings.Repeat("様", 342),
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
			input:   "foobar\nbarfoo",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateTableName(tc.input); (err != nil) != tc.wantErr {
				t.Errorf("ValidateTableName() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
