#!/bin/bash

set -e 

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

( "$SCRIPT_DIR/install_toolchain.sh" )
( "$SCRIPT_DIR/install_workspaces.sh" && make -C e2e test )
