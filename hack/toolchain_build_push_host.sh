#!/bin/bash

set -ex

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}

BUNDLE_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-bundle:${TAG}"
INDEX_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-index:${TAG}"

# checkout
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
git clone --depth 2 https://github.com/codeready-toolchain/toolchain-cicd.git

# build and publish images
make -C registration-service docker-push "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator docker-image "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"

# generate OLM bundle manifests
make -C host-operator bundle "BUNGLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2
(
  # replacing REPLACE_* strings in bundle
  cd host-operator/bundle
  host_image="quay.io/${QUAY_NAMESPACE}/host-operator:${TAG}"
  regsvc_image="quay.io/${QUAY_NAMESPACE}/registration-service:${TAG}"

  sed -i \
    's|REPLACE_IMAGE|'"${host_image}"'|;s|REPLACE_REGISTRATION_SERVICE_IMAGE|'"${regsvc_image}"'|' \
    manifests/toolchain-host-operator.clusterserviceversion.yaml
)

# build OLM images
(
  cd host-operator

  # build bundle
  ${BUILDER} build -f bundle.Dockerfile -t "${BUNDLE_IMAGE}" .

  # build index
  opm index add --permissive --generate --out-dockerfile index.Dockerfile --bundles "${BUNDLE_IMAGE}"
  ${BUILDER} build -f index.Dockerfile -t "${INDEX_IMAGE}" .
)

# push images
make -C host-operator docker-push -o docker-image "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
${BUILDER} push "${BUNDLE_IMAGE}"
${BUILDER} push "${INDEX_IMAGE}"
