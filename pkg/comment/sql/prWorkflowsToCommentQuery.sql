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

WITH recent_closed_pr AS (
    SELECT
        LAX_STRING(payload.pull_request.merge_commit_sha) as merge_commit_sha,
        LAX_STRING(payload.pull_request.head.ref) as head_sha,
        LAX_STRING(payload.pull_request.html_url) as pr_html_url,
        SAFE.INT64(payload.pull_request.id) pr_id,
        received as pr_received
    FROM `unique_events_by_date_type`(@ts_pr_search_start, @ts_pr_search_end, "pull_request")
    WHERE SAFE.BOOL(payload.pull_request.merged) = true AND LAX_STRING(payload.action) = "closed"
),
 already_processed_prs AS (
     SELECT ics.pull_request_id
     FROM  `invocation-comment-status` ics
     GROUP BY ics.pull_request_id
     HAVING COUNT(*) >= 3 OR COUNTIF(ics.status = "SUCCESS") > 0
 ),
 recent_workflow_runs AS (
     SELECT payload as workflow_run_payload, delivery_id
     FROM `unique_events_by_date_type`(@ts_workflow_search_start, @ts_workflow_search_end, "workflow_run")
 ) -- still possible some logs are missed, tradeoff between query cost and completeness


SELECT leech_status.delivery_id as workflow_delivery_id, leech_status.workflow_uri as workflow_uri, leech_status.logs_uri as logs_uri, recent_pr_workflows.pr_id as pr_id, recent_pr_workflows.pr_html_url as pr_url, recent_pr_workflows.workflow_run_payload as workflow_payload, recent_pr_workflows.pr_received as pr_received_time
FROM `leech_status` leech_status
         JOIN (
    SELECT *
    FROM recent_workflow_runs
             JOIN (SELECT recent_closed_pr.*
                   FROM recent_closed_pr
                            LEFT OUTER JOIN `already_processed_prs` appr
                                            ON recent_closed_pr.pr_id = appr.pull_request_id
                   WHERE appr.pull_request_id is null
                   ORDER BY recent_closed_pr.pr_received ASC
                   LIMIT 20) as recent_closed_pr
                  ON recent_closed_pr.merge_commit_sha = SAFE.STRING(recent_workflow_runs.workflow_run_payload.workflow_run.head_sha) OR recent_closed_pr.head_sha = SAFE.STRING(recent_workflow_runs.workflow_run_payload.workflow_run.head_sha)
) AS recent_pr_workflows ON leech_status.delivery_id = recent_pr_workflows.delivery_id
WHERE leech_status.status = "SUCCESS"
