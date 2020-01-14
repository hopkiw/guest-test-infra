# Guest OS Build and Release System

## Overview
Sources for components owned by the Guest OS team are stored in github
repositories within the GoogleCloudPlatform github org. Management of these
sources is done through the prow CI/CD system, which provides OWNERS file
management of github paths, manages github settings such as branch protection,
and handles all merges, tags and builds to the repositories themselves.

### Current repos
Current repos managed by this system are:

* guest-agent
* guest-oslogin
* osconfig
* guest-logging-go

## Prow
Prow is a kubernetes based CI/CD system, consisting of several kubernetes
microservices. We are using the shared OSS cluster run by the kubernetes testing
team, and our job and repo configurations are stored here. To avoid shared
secrets, rather than uploading our service account credentials to the shared
cluster, we run a separate GKE cluster solely for running prow test pods in,
which we've granted the shared cluster permission to do.

## Package build workflow
Aside from use managing development in the github repos, prow builds binary
packages after every merge to master. README.md in this repo describes the
individual container images we have created for these builds, and the prow
config describes the arguments we provide. The general workflow is as follows:

Each repo we support contains its own build configuration in the form of debian
packaging directories, RPM spec files and/or Googet spec files. These files
describe both how to build and how to package the software.

The daisy-builder container generates a new version and provides it to the
relevant packagebuild scripts. The packagebuild scripts use the package build
configuration from the repo to create and upload packages to GCS. The daisy
builder then tags the repo at the specified ref or commit with the generated
version.

The version is specified as `YYYYMMDD.NN`, where YYYYMMDD represents the day of
the build and NN is a build number which starts at 0 and increments with each
build in a given day. The current build number is taken from github version
tags.

## Testing

### Using packagebuild
Using packagebuild to build some software either to produce a release by hand
(rollforward? other uses?) or while actively developing. How to point
packagebuild at your repo.
### Testing packagebuild
When packagebuild itself needs some modifications.
### Using daisy-builder
Emulating a PR or a merge-to-master including version generation and tagging
Using pj-on-kind.sh - Using prow cluster? (is it possible?)
### Emulating daisy-builder
What if you don't have pj-on-kind setup ? Craft a version by hand. Tag by hand.
### Testing daisy-builder
When the daisy-builder container itself needs to be modified.
### Testing prow
Want to develop a new job? Want to run an existing job but against a different
repo or with different arguments? Modifying prow config and running
pj-on-kind.sh

#### Where to get information?

* prow.gflocks.com
* kubectl commands
* locations of logs (thru prow.gflocks, and not)
* locations of artifacts (thru..)
