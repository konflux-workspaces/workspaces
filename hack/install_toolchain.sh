#!/bin/bash

set -e

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

f=$(mktemp --directory /tmp/toolchain.XXXX)

cd "$f"

git clone --depth 2 https://github.com/codeready-toolchain/member-operator.git
git clone --depth 2 --branch public-viewer https://github.com/filariow/toolchain-e2e.git
git clone --depth 2 --branch public-viewer https://github.com/filariow/host-operator.git
git clone --depth 2 --branch public-viewer https://github.com/filariow/toolchain-common.git
git clone --depth 2 --branch public-viewer https://github.com/filariow/toolchain-api.git api
git clone --depth 2 --branch public-viewer https://github.com/filariow/registration-service

make -C toolchain-e2e dev-deploy-e2e-local

oc patch \
  toolchainconfigs.toolchain.dev.openshift.com config \
  -n toolchain-host-operator \
  --patch='{"spec":{"global":{"publicViewer":{"enabled":true,"username":"public-viewer"}}}}' \
  --type=merge

oc delete pods --all -n toolchain-host-operator
