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
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
git clone --depth 2 https://github.com/codeready-toolchain/toolchain-cicd.git

# build and publish images
make -C registration-service docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator bundle "BUNGLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2

make -C host-operator run-cicd-script \
  SCRIPT_PATH=scripts/cd/push-bundle-and-index-image.sh \
  SCRIPT_PARAMS="-pr ../host-operator/ -qn ${QUAY_NAMESPACE} -ch alpha -td /tmp -ib docker -iin host-operator-index -iit ${TAG} -ip linux/amd64 -bt ${TAG}"
