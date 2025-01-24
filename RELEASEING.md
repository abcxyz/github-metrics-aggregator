# Releasing

To create a release for this project follow the following steps:

1. Go to the [Draft Release](https://github.com/abcxyz/github-metrics-aggregator/actions/workflows/draft-release.yml) action and click `Run workflow`
2. Enter the update strategy you would like (`major`, `minor`, `patch`). This project uses [semver](https://semver.org/) for its
   versioning which is where these terms originate.
3. A pull request will then be created with the new version number and release notes. Have this PR reviewed for correctness.
4. Once the PR is approved and merged, a release will be created with the new version and published to GitHub.
