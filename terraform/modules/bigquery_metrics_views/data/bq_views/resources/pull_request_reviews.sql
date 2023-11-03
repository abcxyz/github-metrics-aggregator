-- Copyright 2023 The Authors (see AUTHORS file)
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- Extracts all distinct pull_request_reviews from the pull_request_review_events view.
-- The most recent event for each distinct review id is used to extract
-- the review data. This ensures that the review data is up to date.

SELECT
  pull_request_review_events.body,
  pull_request_review_events.commit_id,
  pull_request_review_events.delivery_id,
  pull_request_review_events.html_url,
  pull_request_review_events.id,
  pull_request_review_events.organization,
  pull_request_review_events.organization_id,
  pull_request_review_events.pull_request_author,
  pull_request_review_events.pull_request_author_id,
  pull_request_review_events.pull_request_id,
  pull_request_review_events.pull_request_number,
  pull_request_review_events.pull_request_url,
  pull_request_review_events.repository,
  pull_request_review_events.repository_full_name,
  pull_request_review_events.repository_id,
  pull_request_review_events.repository_visibility,
  pull_request_review_events.reviewer,
  pull_request_review_events.reviewer_id,
  (pull_request_review_events.reviewer = pull_request_review_events.pull_request_author) AS reviewer_is_author,
  pull_request_review_events.state,
  pull_request_review_events.submitted_at,
FROM
  `${dataset_id}.pull_request_review_events` pull_request_review_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.pull_request_review_events`
  GROUP BY
    id ) unique_pull_request_review_ids
ON
  pull_request_review_events.id = unique_pull_request_review_ids.id
  AND pull_request_review_events.received = unique_pull_request_review_ids.received;
