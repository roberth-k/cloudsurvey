aws codebuild plugins
=====================

# aws_codebuild_builds

#### configuration

- `since` (duration): the amount of time to look back when gathering metrics

#### output

The plugin will produce stats about the builds that completed within the last `since` amount of time.

**name:** `aws_codebuild_build`
**tags:**

- `project_name`: the project name
- `status`: the build status

**fields:**

- `duration` (duration): the total duration of the build
