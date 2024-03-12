#!/bin/bash

set -e -o pipefail +x

_namespace="$1"
_image="$2"

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}

# retrieving toolchain-host namespace
toolchain_host_ns=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
if [[ -z "${toolchain_host_ns}" ]]; then
    toolchain_host_ns="toolchain-host_operator"
fi

# retrieve OCP cluster domain
domain=$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')

# prepare temporary folder
wd=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "${ROOT_DIR}/server/manifests" "${wd}/manifests"

# prepare overlay in the temporary folder
"${DIR}/prepare_overlay.sh" \
  "${wd}/manifests" \
  "local" \
  "${_namespace}" \
  "${_image}" \
  "${toolchain_host_ns}" \
  "apps.${domain}"

# apply manifests
${KUSTOMIZE} build "${wd}/manifests/overlays/local" | ${KUBECLI} apply -f -

# cleanup
rm -r "${wd}"
