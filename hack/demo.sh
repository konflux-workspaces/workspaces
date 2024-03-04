#!/bin/bash

set -e 

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

( ./install_toolchain.sh )
( ./install_workspaces.sh && make -C e2e test )
