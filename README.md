# Compute Engine Guest OS - Test infrastructure

This repository contains tools, configuration and documentation for the public
test infrastructure used by the Google Compute Engine Guest OS team.

## Docker container images

The source for Docker container images used in our
[prow](https://github.com/kubernetes/test-infra/tree/master/prow) jobs is
located under the [container_images](container_images) directory. The
configuration for those jobs lives in a separate repository for the OSS prow
cluster, our config is
[here](https://github.com/GoogleCloudPlatform/oss-test-infra/tree/master/prow/prowjobs/GoogleCloudPlatform/gcp-guest/)

This repository is configured with GCP Cloud Build, such that any commit
triggers new images to be built for every image in this directory.

The following container image definitions exist:

### autoversioner

The `autoversioner` image exists to build the `versiongenerator` and `tagger`
binaries and make them available for consumption by other images, currently only
`daisy-builder`.

### build-essential

The `build-essential` image is a derivative of the latest debian image with
build-essential preinstalled. This is used for generic build tasks.

### daisy-builder

The `daisy-builder` image is meant for building packages. It expects to be run
with this repo checked out, something which prow handles for us automatically.
It uses the tagger and versiongenerator binaries from the autoversioner image,
and generates a dynamic daisy workflow which includes one or more packagebuild
workflows. It takes positional parameters:

1. Build Project ID
1. Build Zone
1. Comma separated list of distros to build
1. GCS output bucket

### gobuild

The `gobuild` image expects to be run with the target repository checked out and
runs `go build` for both Linux and Windows, returning true if both succeed. Used
in prow presubmits.

### gocheck

The `gocheck` image expects to be run with the target repository checked out and
runs `golint` `gofmt` and `go vet` for both Linux and Windows, returning true if
all succeed. Used in prow presubmits.

### gotest

The `gotest` image expects to be run with the target repository checked out and
runs `go test` for both Linux and Windows, returning true if both succeed. Used
in prow presubmits.

### selinux-tools

The `selinux-tools` image contains SELinux-oriented tools, including a custom
build of semodule. This is used currently in presubmits which must validate the
contents of SELinux package files (currently only one is configured, for
guest-oslogin).

## Packagebuilder scripts

The packagebuild directory contains daisy workflows and startup scripts which
build RPM, DEB and Googet packages from a conforming github repository. The
repository to build from must contain a top-level packaging directory containing
one or more of:

* A `debian` directory containing debian build files (changelog, rules, control,
  etc).
* One or more RPM `spec` files.
* A `googet` directory which itself contains:
  * One or more googet `spec` files.
  * Necessary files to build them.

## autoversioner

The autoversioner directory contains the source for the `versiongenerator` and
`tagger` binaries, which are used in our package build and release process.

## License

All files in this repository are under the
[Apache License, Version 2.0](LICENSE) unless noted otherwise.
