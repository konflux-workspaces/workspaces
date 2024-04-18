#!/bin/bash

set -e

# parse input
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}
QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
BUNDLE_IMAGE="quay.io/${QUAY_NAMESPACE}/member-operator-bundle:${TAG}"
INDEX_IMAGE="quay.io/${QUAY_NAMESPACE}/member-operator-index:${TAG}"

# push images
make -C member-operator "${BUILDER}-push" -o "${BUILDER}-image" "QUAY_NAMESPACE=${QUAY_NAMESPACE}" IMAGE_TAG="${TAG}"
${BUILDER} push "${BUNDLE_IMAGE}"
${BUILDER} push "${INDEX_IMAGE}"
