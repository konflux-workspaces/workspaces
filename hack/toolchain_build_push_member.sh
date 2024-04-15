#!/bin/bash

set -e

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-filario}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "${f}"

# checkout
git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 https://github.com/codeready-toolchain/toolchain-cicd.git

# build and publish images
make -C member-operator docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" IMAGE_TAG="${TAG}"

make -C member-operator bundle "BUNDLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2
make -C member-operator run-cicd-script \
  SCRIPT_PATH=scripts/cd/push-bundle-and-index-image.sh \
  SCRIPT_PARAMS="-pr ../member-operator/ -qn ${QUAY_NAMESPACE} -ch alpha -td /tmp -ib docker -iin member-operator-index -iit ${TAG} -ip linux/amd64 -bt ${TAG}"
