# Consolidated Terraform moved blocks

moved {
  from = module.gma.google_cloud_run_v2_job.leech
  to   = module.gma.google_cloud_run_v2_job.artifacts
}

moved {
  from = module.gma.google_cloud_run_v2_job_iam_binding.leech_job_admins
  to   = module.gma.google_cloud_run_v2_job_iam_binding.artifacts_job_admins
}

moved {
  from = module.gma.google_cloud_run_v2_job_iam_binding.leech_job_developers
  to   = module.gma.google_cloud_run_v2_job_iam_binding.artifacts_job_developers
}

moved {
  from = module.gma.google_cloud_run_v2_job_iam_binding.leech_job_invokers
  to   = module.gma.google_cloud_run_v2_job_iam_binding.artifacts_job_invokers
}

moved {
  from = module.gma.google_service_account.leech_sa
  to   = module.gma.google_service_account.artifacts_sa
}

moved {
  from = module.gma.google_project_iam_member.leech_secret_accessor
  to   = module.gma.google_project_iam_member.artifacts_secret_accessor
}

moved {
  from = module.gma.google_project_iam_member.leech_invoker
  to   = module.gma.google_project_iam_member.artifacts_invoker
}

moved {
  from = module.gma.google_project_iam_member.leech_bigquery_job_user
  to   = module.gma.google_project_iam_member.artifacts_bigquery_job_user
}

moved {
  from = module.gma.google_bigquery_dataset_iam_member.leech_dataset_viewer
  to   = module.gma.google_bigquery_dataset_iam_member.artifacts_dataset_viewer
}

moved {
  from = module.gma.google_bigquery_table_iam_member.leech_table_editor
  to   = module.gma.google_bigquery_table_iam_member.artifacts_table_editor
}

moved {
  from = module.gma.google_project_iam_member.leech_storage_object_user
  to   = module.gma.google_project_iam_member.artifacts_storage_object_user
}

moved {
  from = module.gma.google_cloud_scheduler_job.leech_scheduler
  to   = module.gma.google_cloud_scheduler_job.artifacts_scheduler
}

moved {
  from = module.gma.module.leech_alerts
  to   = module.gma.module.artifacts_alerts
}

moved {
  from = module.gma.google_storage_bucket.leech_storage_bucket
  to   = module.gma.google_storage_bucket.artifacts_storage_bucket
}

moved {
  from = module.bigquery.google_bigquery_table.optimized_events_table
  to   = module.bigquery.module.optimized_events_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.optimized_event_owners
  to   = module.bigquery.module.optimized_events_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.optimized_event_editors
  to   = module.bigquery.module.optimized_events_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.optimized_event_viewers
  to   = module.bigquery.module.optimized_events_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.events_table
  to   = module.bigquery.module.events_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.event_owners
  to   = module.bigquery.module.events_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.event_editors
  to   = module.bigquery.module.events_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.event_viewers
  to   = module.bigquery.module.events_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.invocation_comment_table[0]
  to   = module.bigquery.module.invocation_comment_table[0].google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.invocation_comment_owners
  to   = module.bigquery.module.invocation_comment_table[0].google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.invocation_comment_editors
  to   = module.bigquery.module.invocation_comment_table[0].google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.invocation_comment_viewers
  to   = module.bigquery.module.invocation_comment_table[0].google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.leech_table
  to   = module.bigquery.module.artifacts_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.leech_owners
  to   = module.bigquery.module.artifacts_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.leech_editors
  to   = module.bigquery.module.artifacts_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.leech_viewers
  to   = module.bigquery.module.artifacts_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.module.leech_table
  to   = module.bigquery.module.artifacts_table
}

moved {
  from = module.bigquery.google_bigquery_table.prstats
  to   = module.bigquery.module.prstats_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table.checkpoint_table
  to   = module.bigquery.module.checkpoint_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.checkpoint_owners
  to   = module.bigquery.module.checkpoint_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.checkpoint_editors
  to   = module.bigquery.module.checkpoint_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.checkpoint_viewers
  to   = module.bigquery.module.checkpoint_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.commit_review_status_table
  to   = module.bigquery.module.commit_review_status_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.commit_review_status_owners
  to   = module.bigquery.module.commit_review_status_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.commit_review_status_editors
  to   = module.bigquery.module.commit_review_status_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.commit_review_status_viewers
  to   = module.bigquery.module.commit_review_status_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.integration_events
  to   = module.bigquery.module.integration_events_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table.prstats_pull_request_reviews_table
  to   = module.bigquery.module.prstats_pull_request_reviews_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table.prstats_pull_requests_table
  to   = module.bigquery.module.prstats_pull_requests_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table.raw_events_table
  to   = module.bigquery.module.raw_events_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.raw_event_owners
  to   = module.bigquery.module.raw_events_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.raw_event_editors
  to   = module.bigquery.module.raw_events_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.raw_event_viewers
  to   = module.bigquery.module.raw_events_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_bigquery_table.failure_events_table
  to   = module.bigquery.module.failure_events_table.google_bigquery_table.default
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.failure_events_owners
  to   = module.bigquery.module.failure_events_table.google_bigquery_table_iam_member.owners
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.failure_events_editors
  to   = module.bigquery.module.failure_events_table.google_bigquery_table_iam_member.editors
}

moved {
  from = module.bigquery.google_bigquery_table_iam_member.failure_events_viewers
  to   = module.bigquery.module.failure_events_table.google_bigquery_table_iam_member.viewers
}

moved {
  from = module.bigquery.google_pubsub_subscription.relay_optimized_events
  to   = module.pubsub.google_pubsub_subscription.relay_optimized_events
}

moved {
  from = module.gma.google_pubsub_topic.relay
  to   = module.pubsub.google_pubsub_topic.relay
}

moved {
  from = module.gma.google_pubsub_topic_iam_member.relay_topic_remote_subscriber
  to   = module.pubsub.google_pubsub_topic_iam_member.relay_topic_remote_subscriber
}

moved {
  from = module.gma.google_pubsub_topic_iam_member.relay_publisher
  to   = module.pubsub.google_pubsub_topic_iam_member.relay_publisher
}
