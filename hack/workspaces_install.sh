#!/bin/bash

set -e

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}

f=$(mktemp -d /tmp/workspaces-demo.XXXX)

cp -r hack/ operator/ e2e/ server/ "${f}"
cd "${f}"

make -C e2e prepare
