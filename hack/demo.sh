#!/bin/bash

set -e -o pipefail

# parse input
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

CURRENT_DIR="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "${CURRENT_DIR}")"
CI_DIR="$(realpath "$(dirname "${CURRENT_DIR}")/../ci")"

SUFFIX="e2e$(date +'%d%H%M%S')"
echo "using suffix: ${SUFFIX}"

# build and install toolchain
( "${CI_DIR}/toolchain_manager.sh" "publish" "${SUFFIX}" "-n" "${QUAY_NAMESPACE}" && \
  "${CI_DIR}/toolchain_manager.sh" "deploy" "${SUFFIX}" "-n" "${QUAY_NAMESPACE}" )

# build and install workspaces
( "${SCRIPT_DIR}/workspaces_install.sh" && make -C e2e test )
