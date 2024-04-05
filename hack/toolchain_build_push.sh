#!/bin/bash

set -e

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "$f"

# checkout
git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-common.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-api.git api
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-e2e

# build and publish images
make -C member-operator run-cicd-script \
  SCRIPT_PATH=scripts/ci/manage-member-operator.sh \
  SCRIPT_PARAMS="-po true -io false -mn toolchain-member-operator -qn $QUAY_NAMESPACE -ds $TAG -dl false -mr ./"

make -C host-operator run-cicd-script \
  SCRIPT_PATH=scripts/ci/manage-host-operator.sh \
  SCRIPT_PARAMS="-po true -io false -hn toolchain-host-operator -qn $QUAY_NAMESPACE -ds $TAG -dl false -hr ./ -rr ../registration-service"
