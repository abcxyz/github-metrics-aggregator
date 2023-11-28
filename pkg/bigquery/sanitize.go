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

// Start with lowercase, middle is lowercase, number or hyphen, cannot end in
// hyphen. 6-30 characters in length (start, 4-28 middle, end).
const projectIDRegex = `^[a-z][a-z0-9\-]{4,28}[a-z0-9]$`

var projectIDMatcher = regexp.MustCompile(projectIDRegex)

// Lowercase and uppercase letters and underscores. Max 1024 characters.
// regexp only allows 1000 repetitions, so had to manually repeat.
const datasetIDRegex = `^[a-zA-Z0-9_]{1,512}[a-zA-Z0-9_]{0,512}$`

var datasetIDMatcher = regexp.MustCompile(datasetIDRegex)

// Unicode characters in category L (letter), M (mark), N (number),
// Pc (connector, including underscore), Pd (dash), Zs (space). Max 1024
// characters. 1024 is actually UTF-8 bytes from experimentation.
// UTF-8 length check will be done separately.
// regexp only allows 1000 repetitions, so had to manually repeat.
const tableNameRegex = `^[\p{L}\p{M}\p{N}\p{Pc}\p{Pd}\p{Zs}]{1,512}[\p{L}\p{M}\p{N}\p{Pc}\p{Pd}\p{Zs}]{0,512}$`

var tableNameMatcher = regexp.MustCompile(tableNameRegex)

// Returns nil if provided string is a valid GCP project id, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/resource-manager/docs/creating-managing-projects].
// Should allow for safe string interpolation without SQL injection, as
// identifiers cannot be replaced via parameters in BigQuery.
// In most cases, setting the default GCP project in the client's Query object
// is preferable as it avoids the injection risk entirely.
// Does not check for restricted strings such as google, null, ect.
func ValidateGCPProjectID(projectID string) error {
	if !projectIDMatcher.MatchString(projectID) {
		return fmt.Errorf("invalid GCP project id")
	}
	return nil
}

// Returns nil if provided string is a valid dataset id, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/bigquery/docs/datasets#dataset-naming].
// Should allow for safe string interpolation without SQL injection, as
// identifiers cannot be replaced via parameters in BigQuery.
// In most cases, setting the default dataset in the client's Query object
// is preferable as it avoids the injection risk entirely.
// Does not check for restricted strings such as google, null, ect.
func ValidateDatasetID(datasetID string) error {
	if !datasetIDMatcher.MatchString(datasetID) {
		return fmt.Errorf("invalid dataset id")
	}
	return nil
}

// Returns nil if provided string is a valid table name, else returns an
// error. Based on rules defined in
// [https://cloud.google.com/bigquery/docs/tables#table_naming].
// Should allow for safe string interpolation without SQL injection, as
// identifiers cannot be replaced via parameters in BigQuery.
// Hardcoding table names may be preferred to avoid any injection risk.
// Unclear if these rules are valid for all external tables.
// Does not check for restricted strings such as google, null, ect.
func ValidateTableName(tableName string) error {
	// TODO: is this really necessary, seems like regexp may do this for free.
	if !utf8.Valid([]byte(tableName)) {
		return fmt.Errorf("invalid table name: not UTF-8")
	}
	// Checking to ensure max 1024 UTF-8 bytes, as that is the limit actually
	// in place.
	if len(tableName) > 1024 {
		return fmt.Errorf("invalid table name: too many bytes")
	}
	// Regex has some length validation, though only lower bound should ever be
	// hit.
	if !tableNameMatcher.MatchString(tableName) {
		return fmt.Errorf("invalid table name")
	}
	return nil
}
