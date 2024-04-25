#!/bin/bash

# Builds and publishes kubesaw components to a quay repository.  Used as a part
# of the CI process for testing in openshift-ci.

set -e

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
TAG=${1:-e2etest}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "${f}"

# clone
git clone --depth=2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/toolchain-e2e.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/registration-service

# build & publish
make -C toolchain-e2e publish-current-bundles-for-e2e \
    FORCED_TAG="${TAG}" \
    REG_REPO_PATH=../registration-service \
    HOST_REPO_PATH=../host-operator \
    MEMBER_REPO_PATH=../member-operator \
    CI=true
