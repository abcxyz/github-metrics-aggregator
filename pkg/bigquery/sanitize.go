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

// Package bigquery implements helpers for using BigQuery.
package bigquery

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

const tableNameMaxUTF8Bytes = 1024

// Start with lowercase, middle is lowercase, number or hyphen, cannot end in
// hyphen. 6-30 characters in length (start, 4-28 middle, end).
var projectIDMatcher = regexp.MustCompile(`\A[a-z][a-z0-9\-]{4,28}[a-z0-9]\z`)

// Lowercase and uppercase letters and underscores. Max 1024 characters.
// regexp only allows 1000 repetitions, so had to manually repeat.
const datasetIDRegex = `\A[a-zA-Z0-9_]{1,512}[a-zA-Z0-9_]{0,512}\z`

var datasetIDMatcher = regexp.MustCompile(datasetIDRegex)

// Unicode characters in category L (letter), M (mark), N (number),
// Pc (connector, including underscore), Pd (dash), Zs (space). Max 1024
// characters. 1024 is actually UTF-8 bytes from experimentation.
// UTF-8 length check will be done separately.
// regexp only allows 1000 repetitions, so had to manually repeat.
const tableNameRegex = `\A[\p{L}\p{M}\p{N}\p{Pc}\p{Pd}\p{Zs}]+\z`

var tableNameMatcher = regexp.MustCompile(tableNameRegex)

// Returns nil if provided string is a valid GCP project id, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/resource-manager/docs/creating-managing-projects].
// Identifiers cannot be replaced via parameters in BigQuery, so interpolation
// is a necessry evil.
// In most cases, setting the default GCP project in the client's Query object
// is preferable as it avoids the injection risk entirely.
// Does not check for restricted strings such as google, null, etc.
func ValidateGCPProjectID(projectID string) error {
	if !projectIDMatcher.MatchString(projectID) {
		return fmt.Errorf("invalid GCP project id")
	}
	return nil
}

// Returns nil if provided string is a valid dataset id, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/bigquery/docs/datasets#dataset-naming].
// Identifiers cannot be replaced via parameters in BigQuery, so interpolation
// is a necessry evil.
// In most cases, setting the default dataset in the client's Query object
// is preferable as it avoids the injection risk entirely.
// Does not check for restricted strings such as google, null, etc.
func ValidateDatasetID(datasetID string) error {
	if !datasetIDMatcher.MatchString(datasetID) {
		return fmt.Errorf("invalid dataset id")
	}
	return nil
}

// Returns nil if provided string is a valid table name, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/bigquery/docs/tables#table_naming].
// Identifiers cannot be replaced via parameters in BigQuery, so interpolation
// is a necessry evil.
// Hardcoding table names may be preferred to avoid any injection risk.
// Unclear if these rules are valid for all external tables.
// Does not check for restricted strings such as google, null, etc.
func ValidateTableName(tableName string) error {
	if !utf8.Valid([]byte(tableName)) {
		return fmt.Errorf("invalid table name: not UTF-8")
	}
	// Checking to ensure max UTF-8 bytes, as that is the limit actually
	// in place.
	if l := len(tableName); l > tableNameMaxUTF8Bytes || l < 1 {
		return fmt.Errorf("invalid table name: too few/many bytes: got %d expected [1, %v]", l, tableNameMaxUTF8Bytes)
	}
	// Regex has some length validation, though only lower bound should ever be
	// hit.
	if !tableNameMatcher.MatchString(tableName) {
		return fmt.Errorf("invalid table name")
	}
	return nil
}
