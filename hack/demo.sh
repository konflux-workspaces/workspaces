#!/bin/bash

set -e -o pipefail

# parse input
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

CURRENT_DIR="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "${CURRENT_DIR}")"

SUFFIX="e2e$(date +'%d%H%M%S')"
echo "using suffix: ${SUFFIX}"

( 
  # create a temporary directory
  f=$(mktemp --directory /tmp/toolchain.XXXX)
  cd "${f}"
  
  # checkout repos
  "${SCRIPT_DIR}/toolchain_host_checkout.sh" "${SUFFIX}"
  "${SCRIPT_DIR}/toolchain_member_checkout.sh" "${SUFFIX}"

  # build images
  "${SCRIPT_DIR}/toolchain_host_build.sh" "${SUFFIX}" 
  "${SCRIPT_DIR}/toolchain_member_build.sh" "${SUFFIX}" 
  
  # push images
  "${SCRIPT_DIR}/toolchain_host_push.sh" "${SUFFIX}"
  "${SCRIPT_DIR}/toolchain_member_push.sh" "${SUFFIX}"
)

# install toolchain
( "${SCRIPT_DIR}/toolchain_install.sh" "${SUFFIX}" )

# build and install workspaces
( "${SCRIPT_DIR}/workspaces_install.sh" && make -C e2e test )
