#!/bin/bash

set -e

# parse input
HOST_BRANCH=${HOST_BRANCH:-pv-531-sbcleanup}
E2E_BRANCH=${E2E_BRANCH:-pv-532-proxy}
REGSVC_BRANCH=${REGSVC_BRANCH:-pv-532-proxy}

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
TAG=${1:-e2etest}
KUBECLI=${KUBECLI:-kubectl}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "${f}"

# checkout
git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 --branch "${E2E_BRANCH}" https://github.com/filariow/toolchain-e2e.git
git clone --depth 2 --branch "${HOST_BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 2 --branch "${REGSVC_BRANCH}" https://github.com/filariow/registration-service

# deploy
(
  set -e -o pipefail

  ${KUBECLI} create namespace toolchain-host-operator --dry-run=client --output=yaml | \
    ${KUBECLI} apply -f -
  ${KUBECLI} create namespace toolchain-member-operator --dry-run=client --output=yaml | \
    ${KUBECLI} apply -f -

  cd "${f}/toolchain-e2e"

  make dev-deploy-e2e-local PUBLISH_OPERATOR=true DATE_SUFFIX="${TAG}" DEPLOY_LATEST=false
)

# patch configuration
${KUBECLI} patch \
  toolchainconfigs.toolchain.dev.openshift.com config \
  -n toolchain-host-operator \
  --patch='{"spec":{"publicViewerConfig":{"enabled":true}}}' \
  --type=merge

# restart operator
${KUBECLI} delete pods --all -n toolchain-host-operator
