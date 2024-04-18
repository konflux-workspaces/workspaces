#!/bin/bash

set -e

# parse input
BUILDER=${BUILDER:-docker}
TAG=${1:-e2etest}
QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
BUNDLE_IMAGE="quay.io/${QUAY_NAMESPACE}/member-operator-bundle:${TAG}"
INDEX_IMAGE="quay.io/${QUAY_NAMESPACE}/member-operator-index:${TAG}"

# build images
make -C member-operator ${BUILDER}-image "QUAY_NAMESPACE=${QUAY_NAMESPACE}" IMAGE_TAG="${TAG}"

# generate OLM bundle manifests
make -C member-operator bundle "BUNDLE_TAG=${TAG}" CHANNEL=alpha NEXT_VERSION=0.0.2
(
  # replacing REPLACE_* strings in bundle
  cd member-operator/bundle
  member_image="quay.io/${QUAY_NAMESPACE}/member-operator:${TAG}"
  wcp_image="quay.io/${QUAY_NAMESPACE}/member-operator-console-plugin:${TAG}"
  webhook_image="quay.io/${QUAY_NAMESPACE}/member-operator-webhook:${TAG}"

  sed -i \
    's|REPLACE_IMAGE|'"${member_image}"'|;s|REPLACE_MEMBER_OPERATOR_WEBCONSOLEPLUGIN_IMAGE|'"${wcp_image}"'|;s|REPLACE_MEMBER_OPERATOR_WEBHOOK_IMAGE|'"${webhook_image}"'|' \
    manifests/toolchain-member-operator.clusterserviceversion.yaml
)

# build OLM images
(
  cd member-operator
  # build bundle
  ${BUILDER} build -f bundle.Dockerfile -t "${BUNDLE_IMAGE}" .
  
  # Needed for building correctly the bundle
  ${BUILDER} push "${BUNDLE_IMAGE}"

  # build index
  opm index add\
    --generate\
    --out-dockerfile index.Dockerfile\
    --bundles "${BUNDLE_IMAGE}"\
    --build-tool "${BUILDER}"\
    --tag "${INDEX_IMAGE}"\
    --pull-tool "${BUILDER}"

  ${BUILDER} build -f index.Dockerfile -t "${INDEX_IMAGE}" .
)
