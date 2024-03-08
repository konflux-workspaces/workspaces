#!/bin/bash

set -e -o pipefail

ROOT_DIR="$(realpath "$(dirname "$(readlink -f "$0")")/../..")"
KUBECLI=${KUBECLI:-kubectl}

# retrieving toolchain-host namespace
toolchain_host=$($KUBECLI get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)

# prepare temporary folder
f=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "$ROOT_DIR/server/manifests" "$f/manifests"

# updating manifests locally
cd "$f/manifests/default"
$KUSTOMIZE edit set namespace "$1"
$KUSTOMIZE edit set image workspaces/rest-api="$2"
$KUSTOMIZE edit add configmap rest-api-server-config \
        --behavior=replace \
        --from-literal=kubesaw.namespace="$( [[ -n "$toolchain_host" ]] && echo "$toolchain_host" || echo "toolchain-host-operator" )"

# apply manifests
$KUSTOMIZE build . | $KUBECLI apply -f -

# cleanup
rm -r "$f"
