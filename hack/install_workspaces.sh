#!/bin/bash

set -e

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-filario}

f=$(mktemp --directory /tmp/workspaces-demo.XXXX)

cp -r operator/ e2e/ server/ "$f"
cd "$f" 

make -C e2e prepare
