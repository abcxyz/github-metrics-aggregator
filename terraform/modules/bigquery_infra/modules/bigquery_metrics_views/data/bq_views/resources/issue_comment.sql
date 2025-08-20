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
-- Extracts all distinct pull_requests from the issue_comments.
-- The most recent event for each distinct comment id is used to extract
-- the comment data. This ensures that the comment data is up to date.

SELECT
  issue_comment_events.organization,
  issue_comment_events.organization_id,
  issue_comment_events.repository,
  issue_comment_events.repository_id,
  issue_comment_events.body,
  issue_comment_events.commenter_id,
  issue_comment_events.commenter,
  issue_comment_events.created_at,
  issue_comment_events.delivery_id,
  issue_comment_events.html_url,
  issue_comment_events.id,
  issue_comment_events.issue_id,
  issue_comment_events.line,
  issue_comment_events.path,
  issue_comment_events.updated_at,
FROM
  `${dataset_id}.issue_comment_events` issue_comment_events
JOIN (
  SELECT id, max(received) received
  FROM `${dataset_id}.issue_comment_events`
  GROUP BY id
) unique_issue_comment_ids
ON issue_comment_events.id = unique_issue_comment_ids.id
AND issue_comment_events.received = unique_issue_comment_ids.received;
