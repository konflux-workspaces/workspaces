#!/bin/bash

set -e -o pipefail +x

_namespace="$1"

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
HELM=${HELM:-helm}

# retrieving toolchain-host namespace
toolchain_host_ns=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
if [[ -z "${toolchain_host_ns}" ]]; then
    toolchain_host_ns="toolchain-host_operator"
fi

toolchain_api_route=$(${KUBECLI} get route -n "${toolchain_host_ns}" api -o jsonpath='{.status.ingress[0].host}')

# retrieve OCP cluster domain
domain=$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')

# deploy with Helm
helm_folder="${ROOT_DIR}/ingress/helm"
${HELM} dependency build "$helm_folder" 
${HELM} upgrade --install workspaces-ingress-internal "$helm_folder" \
  --namespace "$_namespace" --create-namespace \
  --set providers.kubernetesIngress.namespaces="$_namespace" \
  --set host="workspaces-api.apps.${domain}" \
  --set kubesaw.url="${toolchain_api_route}"
