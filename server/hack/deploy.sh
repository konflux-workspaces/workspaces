#!/bin/bash

set -e -o pipefail

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}

# retrieving toolchain-host namespace
toolchain_host=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
if [[ -z "${toolchain_host}" ]]; then
    toolchain_host="toolchain-host_operator"
fi

# prepare temporary folder
f=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "${ROOT_DIR}/server/manifests" "${f}/manifests"

# updating manifests locally
cd "${f}/manifests/default"
${KUSTOMIZE} edit set namespace "$1"
${KUSTOMIZE} edit set image workspaces/rest-api="$2"
${KUSTOMIZE} edit add configmap rest-api-server-config \
        --behavior=replace \
        --from-literal=kubesaw.namespace="${toolchain_host}"

# apply manifests
${KUSTOMIZE} build . | ${KUBECLI} apply -f -

# cleanup
rm -r "${f}"
