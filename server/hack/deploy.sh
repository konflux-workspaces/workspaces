#!/bin/bash

set -e -o pipefail

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}
YQ=${YQ:-yq}

# retrieve toolchain-host namespace
#
# if the namespace doesn't exist, then head will return a failure (since it has
# an empty input).  Since we manually check the error case, disable pipefail
set +o pipefail
toolchain_host=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
if [[ -z "${toolchain_host}" ]]; then
    toolchain_host="toolchain-host_operator"
fi
set -o pipefail

# prepare temporary folder
f=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "${ROOT_DIR}/server/manifests" "${f}/manifests"
cd "${f}/manifests/default"

# updating JWT configuration
if [[ -n "${JWKS_URL}" ]]; then
  ${YQ} eval \
    '.authSources.jwtSource.jwt.jwksUrl = "'"${JWKS_URL}"'"' \
    --inplace "${f}/manifests/server/proxy-config/traefik.yaml"
else
  # add secret to manifests
  private_key=$(openssl genrsa 2048)
  public_key=$(echo "${private_key}" | openssl rsa -pubout 2>/dev/null )

  ${KUSTOMIZE} edit add secret traefik-jwt-keys \
    --disableNameSuffixHash \
    --from-literal=public="${public_key}" \
    --from-literal=private="${private_key}"

  # update traefik config
  ${YQ} eval \
    '.http.middlewares.jwt-authorizer.plugin.jwt.keys[0]="'"${public_key}"'"' \
    --inplace "${f}/manifests/server/proxy-config/dynamic/config.yaml"
fi

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
