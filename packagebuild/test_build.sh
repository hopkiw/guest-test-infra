#!/bin/bash

# Produce a single test build.

[[ -z $TYPE ]] && read -p "Build type [deb9]: " TYPE
[[ -z $PROJECT ]] && read -p "Build project: " PROJECT
[[ -z $ZONE ]] && read -p "Build zone [us-west1-a]: " ZONE
[[ -z $OWNER ]] && read -p "Repo owner or org: " OWNER
[[ -z $REPO ]] && read -p "Repo name: " REPO
[[ -z $GIT_REF ]] && read -p "Ref [master]: " GIT_REF

[[ $TYPE == "" ]] && TYPE="deb9"
[[ $ZONE == "" ]] && ZONE="us-west1-a"
[[ $GIT_REF == "" ]] && GIT_REF="master"

WF="workflows/build_${TYPE}.wf.json"

if [[ ! -f "$WF" ]]; then
  echo "Unknown build type $TYPE"
  exit 1
fi

daisy \
  -project $PROJECT \
  -zone $ZONE \
  -var:gcs_path='${SCRATCHPATH}/packages' \
  -var:repo_owner=$OWNER \
  -var:repo_name=$REPO \
  -var:git_ref=$GIT_REF \
  -var:version=1dummy \
  "$WF"
