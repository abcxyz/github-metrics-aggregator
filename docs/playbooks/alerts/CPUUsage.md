# CPUUsage

## High CPU Utilization

This alert fires when a Cloud Run service or job is experiencing high CPU utilization across all container instances. The alert policy monitors the P99 sampling metric. We consider any lower p value to be too relaxed. 

When 99% or lower of all containers allocated to the service or job reach a configurable threshold percentage, this alert will fire. In other words if the service has 100 instances and at least 2 of those instances are observing 80% CPU utilization or higher the P99 will trigger.

### Triage Steps

1. Observe the metric over a longer period of time (days or weeks) to see if the CPU utilization has been steadily increasing.
2. Review recent deployments to see if a recent change may have caused the increase.
