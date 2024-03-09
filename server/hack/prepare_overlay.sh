#!/bin/bash

set -e -o pipefail

# parsing positional arguments
[[ "$#" -ne "6" ]] && echo "expected 6 arguments, found $#" && exit 1
_manifests_folder="$1"
_overlay="$2"
_namespace="$3"
_image="$4"
_toolchain_host_ns="$5"
_domain="$6"

# tooling
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}

# adding local overlay
overlay_folder="$(realpath "${_manifests_folder}/overlays")/${_overlay}"
mkdir -p "${overlay_folder}"
cd "${overlay_folder}"

${KUSTOMIZE} create
${KUSTOMIZE} edit add base "../../default"
${KUSTOMIZE} edit set namespace "${_namespace}"
${KUSTOMIZE} edit set image workspaces/rest-api="${_image}"
${KUSTOMIZE} edit add configmap rest-api-server-config \
    --behavior=replace \
    --from-literal=kubesaw.namespace="${_toolchain_host_ns}"

## patch ingress
cat << EOF > 'patch-ingress.yaml'
- op: replace
  path: /spec/rules/0/host
  value: workspaces-ingress-proxy-${_namespace}.${_domain}
EOF

${KUSTOMIZE} edit add patch \
  --path='patch-ingress.yaml' \
  --group='networking.k8s.io' \
  --version='v1' \
  --kind='Ingress' \
  --name='workspaces-api'
