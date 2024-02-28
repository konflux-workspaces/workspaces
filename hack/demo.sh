#!/bin/bash

set -e 

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-filario}

( ./install_toolchain.sh )
( ./install_workspaces.sh && make -C e2e test )
