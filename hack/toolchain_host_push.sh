#!/bin/bash

set -ex

# parse input
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}
QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
BUNDLE_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-bundle:${TAG}"
INDEX_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-index:${TAG}"

# push images
make -C registration-service "${BUILDER}-push" -o "${BUILDER}-image" "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator "${BUILDER}-push" -o "${BUILDER}-image" "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
${BUILDER} push "${BUNDLE_IMAGE}"
${BUILDER} push "${INDEX_IMAGE}"
