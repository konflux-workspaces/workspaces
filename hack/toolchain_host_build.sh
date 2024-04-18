#!/bin/bash

set -ex

# parse input
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}
QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
BUNDLE_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-bundle:${TAG}"
INDEX_IMAGE="quay.io/${QUAY_NAMESPACE}/host-operator-index:${TAG}"

# build and publish images
make -C registration-service ${BUILDER}-image "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"
make -C host-operator ${BUILDER}-image "QUAY_NAMESPACE=${QUAY_NAMESPACE}" "IMAGE_TAG=${TAG}"

# generate OLM bundle manifests
make -C host-operator bundle "BUNDLE_TAG=${TAG}" CHANNEL=alpha
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
  
  # Needed for building correctly the index image
  ${BUILDER} push "${BUNDLE_IMAGE}"

  # build index
  opm index add\
    --generate\
    --out-dockerfile index.Dockerfile\
    --bundles "${BUNDLE_IMAGE}"\
    --container-tool "${BUILDER}"\
    --tag "${INDEX_IMAGE}"

  ${BUILDER} build -f index.Dockerfile -t "${INDEX_IMAGE}" .
)
