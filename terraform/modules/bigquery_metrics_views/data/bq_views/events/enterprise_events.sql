
-- Copyright 2025 The Authors (see AUTHORS file)
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

-- Extracts all events containing enterprise information from the raw_events table.
-- This view serves as the base for the final `enterprises` resource view.

SELECT
  SAFE.INT64(payload.enterprise.id) AS id
FROM
  `${raw_events_table_id}`
WHERE
  -- Filter for events that contain enterprise information
  payload.enterprise.id IS NOT NULL;
