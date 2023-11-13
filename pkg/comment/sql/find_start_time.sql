WITH retryable_errors AS (
    SELECT MIN(pull_request_merge_event_received) AS event_time
    FROM invocation_comment_status
    GROUP BY pull_request_id
    HAVING COUNT(*) < 3 AND COUNTIF(status = "SUCCESS") = 0
), most_recent_success as (
    SELECT MAX(pull_request_merge_event_received) AS event_time
    FROM invocation_comment_status
    WHERE status = "SUCCESS"
)
SELECT MAX(event_time) as start_time FROM (
    SELECT MIN(event_time) as event_time FROM (
          SELECT MIN(event_time) FROM retryable_errors
          UNION ALL
          SELECT event_time FROM most_recent_success
      )
    UNION ALL
    SELECT TIMESTAMP(@earliest_timestamp)
);
