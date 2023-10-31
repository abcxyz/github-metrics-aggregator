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
-- Extracts all distinct pull_requests from the pull_request_events_view.
-- The most recent event for each distinct pull request id is used to extract
-- the pull request data. This ensures that the pull request data is up to date.

SELECT
  pull_request_events.active_lock_reason,
  pull_request_events.additions,
  pull_request_events.author,
  pull_request_events.author_id,
  pull_request_events.base_ref,
  pull_request_events.changed_files,
  pull_request_events.closed_at,
  pull_request_events.comments,
  pull_request_events.commits,
  pull_request_events.created_at,
  pull_request_events.deletions,
  pull_request_events.draft,
  pull_request_events.head_ref,
  pull_request_events.html_url,
  pull_request_events.id,
  pull_request_events.locked,
  pull_request_events.maintainer_can_modify,
  pull_request_events.merge_commit_sha,
  pull_request_events.mergeable_state,
  pull_request_events.merged,
  pull_request_events.merged_at,
  pull_request_events.merged_by,
  pull_request_events.number,
  pull_request_events.open_duration_seconds,
  pull_request_events.organization,
  pull_request_events.organization_id,
  pull_request_events.repository,
  pull_request_events.repository_full_name,
  pull_request_events.repository_id,
  pull_request_events.repository_visibility,
  pull_request_events.state,
  pull_request_events.title,
  pull_request_events.updated_at,
  (CASE
      WHEN pull_request_events.additions + pull_request_events.deletions <= 9 THEN 'XS'
      WHEN pull_request_events.additions + pull_request_events.deletions <= 49 THEN 'S'
      WHEN pull_request_events.additions + pull_request_events.deletions <= 249 THEN 'M'
      WHEN pull_request_events.additions + pull_request_events.deletions <= 999 THEN 'L'
    ELSE
    'XL'
  END
    ) AS pr_size,
  (CASE
      WHEN pull_request_events.open_duration_seconds < 600 THEN '< 10m'
      WHEN pull_request_events.open_duration_seconds < 1800 THEN '< 30m'
      WHEN pull_request_events.open_duration_seconds < 3600 THEN '< 1h'
      WHEN pull_request_events.open_duration_seconds < 10800 THEN '< 3h'
      WHEN pull_request_events.open_duration_seconds < 21600 THEN '< 6h'
      WHEN pull_request_events.open_duration_seconds < 43200 THEN '< 12h'
      WHEN pull_request_events.open_duration_seconds < 86400 THEN '< 1d'
      WHEN pull_request_events.open_duration_seconds < 172800 THEN '< 2d'
      WHEN pull_request_events.open_duration_seconds < 345600 THEN '< 4d'
      WHEN pull_request_events.open_duration_seconds < 604800 THEN '< 7d'
      WHEN pull_request_events.open_duration_seconds < 1209600 THEN '< 14d'
      WHEN pull_request_events.open_duration_seconds < 2592000 THEN '< 30d'
    ELSE
    '>= 30d'
  END
    ) AS submission_time,
FROM
  `${dataset_id}.pull_request_events` pull_request_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.pull_request_events`
  GROUP BY
    id ) unique_pull_request_ids
ON
  pull_request_events.id = unique_pull_request_ids.id
  AND pull_request_events.received = unique_pull_request_ids.received;
