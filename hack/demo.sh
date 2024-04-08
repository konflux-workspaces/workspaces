#!/bin/bash

set -e -o pipefail

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

CURRENT_DIR="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "${CURRENT_DIR}")"

( "${SCRIPT_DIR}/install_toolchain.sh" )
( "${SCRIPT_DIR}/install_workspaces.sh" && make -C e2e test )
