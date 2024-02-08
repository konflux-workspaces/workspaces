#!/bin/bash

set -e

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

f=$(mktemp --directory /tmp/toolchain.XXXX)

cd "$f"

git clone --depth 2 git@github.com:codeready-toolchain/member-operator.git
git clone --depth 2 git@github.com:codeready-toolchain/toolchain-e2e.git
git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/host-operator.git
git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/toolchain-common.git
git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/toolchain-api.git api
git clone --depth 2 --branch f45-demo-workspaces git@github.com:filariow/registration-service

make -C toolchain-e2e dev-deploy-e2e-local

