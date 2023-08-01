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
-- Extracts all distinct releases from the release_events view.
-- The most recent event for each distinct release id is used to extract
-- the release data. This ensures that the release data is up to date.

SELECT
  releases_events.author,
  releases_events.author_id,
  releases_events.body,
  releases_events.created_at,
  releases_events.draft,
  releases_events.html_url,
  releases_events.id,
  releases_events.name,
  releases_events.prerelease,
  releases_events.tag_name,
  releases_events.tarball_url,
  releases_events.target_commitish,
  releases_events.upload_url,
  releases_events.zipball_url,
FROM
  `${dataset_id}.releases_events` releases_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.releases_events`
  GROUP BY
    id ) unique_release_ids
ON
  releases_events.id = unique_release_ids.id
  AND releases_events.received = unique_release_ids.received;
