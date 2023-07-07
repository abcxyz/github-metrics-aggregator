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

-- Extracts all distinct pull_request_reviews from the pull_request_review_comment_events view.
-- The most recent event for each distinct comment id is used to extract
-- the comment data. This ensures that the comment data is up to date.

SELECT
  pull_request_review_comment_events.body,
  pull_request_review_comment_events.commenter,
  pull_request_review_comment_events.commenter_id,
  pull_request_review_comment_events.commit_sha,
  pull_request_review_comment_events.created_at,
  pull_request_review_comment_events.diff_hunk,
  pull_request_review_comment_events.html_url,
  pull_request_review_comment_events.id,
  pull_request_review_comment_events.in_reply_to_id,
  pull_request_review_comment_events.line,
  pull_request_review_comment_events.organization,
  pull_request_review_comment_events.organization_id,
  pull_request_review_comment_events.path,
  pull_request_review_comment_events.pull_request_id,
  pull_request_review_comment_events.pull_request_review_id,
  pull_request_review_comment_events.repository,
  pull_request_review_comment_events.repository_id,
FROM
  `${dataset_id}.pull_request_review_comment_events` pull_request_review_comment_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.pull_request_review_comment_events`
  GROUP BY
    id ) unique_pull_request_review_comment_ids
ON
  pull_request_review_comment_events.id = unique_pull_request_review_comment_ids.id
  AND pull_request_review_comment_events.received = unique_pull_request_review_comment_ids.received;
