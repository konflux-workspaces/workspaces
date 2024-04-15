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
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-common.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-api.git api
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/toolchain-e2e

# build and publish images
make -C member-operator docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" IMAGE_TAG="${TAG}"
make -C member-operator bundle push-bundle-and-index-image "BUNDLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2

make -C registration-service docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator bundle push-bundle-and-index-image "BUNGLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2
