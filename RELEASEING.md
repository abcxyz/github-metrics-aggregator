# Releasing

To create a release for this project follow the following steps:

1. Run the `create-tag` workflow and give it a new tag e.g. `v1.0.0`.
2. Get a reviewer to approve the tag creation.
3. Watch for a deployment to `production` and approve it. This is the promotion of the CI container image to the release GAR environment.
