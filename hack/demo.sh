#!/bin/bash

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-filario}

f=$(mktemp --directory /tmp/workspaces-demo.XXXX)

cp -r operator/ e2e/ "$f"
cd "$f" || exit

(
  set -ex

  mkdir toolchain
  cd toolchain || exit 1
  ( git clone --depth 2 git@github.com:codeready-toolchain/member-operator.git && \
      git clone --depth 2 git@github.com:codeready-toolchain/toolchain-e2e.git && \
      git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/host-operator.git && \
      git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/toolchain-common.git && \
      git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/toolchain-api.git api && \
      git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/registration-service ) || exit 1

  make -C toolchain-e2e dev-deploy-e2e-local
) || exit $?

make -C e2e prepare test
