#!/bin/bash

set -e

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "${f}"

# checkout
git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 https://github.com/codeready-toolchain/toolchain-cicd.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-common.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-api.git api
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-e2e

# build and publish images
make -C member-operator docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" IMAGE_TAG="${TAG}"

make -C host-operator run-cicd-script \
  SCRIPT_PATH=scripts/ci/manage-host-operator.sh \
  SCRIPT_PARAMS="-po true -io false -hn toolchain-host-operator -qn ${QUAY_NAMESPACE} -ds ${TAG} -dl false -hr ./ -rr ../registration-service"
