#!/bin/bash

set -e

export QUAY_NAMESPACE=${QUAY_NAMESPACE:-workspaces}

f="$(pwd)"
[[ "${f}" == "/tmp/*" ]] || {
  f=$(mktemp --directory /tmp/workspaces-demo.XXXX)
  cp -r hack/ operator/ e2e/ server/ ingress/ "$f" 
}

make -C "$f/e2e" prepare
