#!/bin/bash

set -e -o pipefail

# parse input
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

CURRENT_DIR="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "${CURRENT_DIR}")"

SUFFIX="e2e$(date +'%d%H%M%S')"
echo "using suffix: ${SUFFIX}"

# build and install toolchain
( "${SCRIPT_DIR}/toolchain_build_push.sh" "${SUFFIX}" )
( "${SCRIPT_DIR}/toolchain_install.sh" "${SUFFIX}" )

# build and install workspaces
( "${SCRIPT_DIR}/workspaces_install.sh" && make -C e2e test )
