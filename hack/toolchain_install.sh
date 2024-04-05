#!/bin/bash

set -e

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}
TAG=${1:-e2etest}

# create a temporary direction
f=$(mktemp --directory /tmp/toolchain.XXXX)
cd "$f"

# checkout
git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/toolchain-e2e.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/toolchain-common.git
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/toolchain-api.git api
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/registration-service

# deploy
( 
  kubectl create namespace toolchain-host-operator --dry-run=client --output=yaml | \
    kubectl apply -f -
  kubectl create namespace toolchain-member-operator --dry-run=client --output=yaml | \
    kubectl apply -f -

  cd "$f/toolchain-e2e"

  make dev-deploy-e2e-local PUBLISH_OPERATOR=false DATE_SUFFIX="$TAG" DEPLOY_LATEST=false
)

# patch configuration
oc patch \
  toolchainconfigs.toolchain.dev.openshift.com config \
  -n toolchain-host-operator \
  --patch='{"spec":{"global":{"publicViewer":{"enabled":true,"username":"public-viewer"}}}}' \
  --type=merge

# restart operator
oc delete pods --all -n toolchain-host-operator
