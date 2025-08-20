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
-- Extracts all distinct teams from the team_events view.
-- The most recent event for each distinct team id is used to extract
-- the team data. This ensures that the team data is up to date.

SELECT
  team_events.deleted,
  team_events.delivery_id,
  team_events.description,
  team_events.html_url,
  team_events.id,
  team_events.members_url,
  team_events.name,
  team_events.organization,
  team_events.organization_id,
  team_events.parent_id,
  team_events.parent_name,
  team_events.permission,
  team_events.privacy,
  team_events.repository,
  team_events.repository_id,
  team_events.notification_setting,
  team_events.slug,
FROM
  `${dataset_id}.team_events` team_events
JOIN(
  SELECT id, MAX(received) received
  FROM `${dataset_id}.team_events`
  GROUP BY id
) unique_team_ids
ON team_events.id = unique_team_ids.id
AND team_events.received = unique_team_ids.received;
