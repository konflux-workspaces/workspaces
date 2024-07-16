#!/bin/bash

set -e -o pipefail

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}
YQ=${YQ:-yq}
SERVER_LOG_LEVEL=${SERVER_LOG_LEVEL:-0}

# retrieve toolchain-host namespace
#
# if the namespace doesn't exist, then head will return a failure (since it has
# an empty input).  Since we manually check the error case, disable pipefail
if [[ -z "${TOOLCHAIN_HOST}" ]]; then
    set +o pipefail
    TOOLCHAIN_HOST=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
    if [[ -z "${TOOLCHAIN_HOST}" ]]; then
        TOOLCHAIN_HOST="toolchain-host-operator"
    fi
    set -o pipefail
fi

# prepare temporary folder
f=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "${ROOT_DIR}/server/config" "${f}/config"
cd "${f}/config/default"

# updating JWT configuration
if [[ -n "${JWKS_URL}" ]]; then
  ${YQ} eval \
    '.authSources.jwtSource.jwt.jwksUrl = "'"${JWKS_URL}"'"' \
    --inplace "${f}/config/server/proxy-config/traefik.yaml"
else
  # add secret to config
  private_key=$(openssl genrsa 2048)
  public_key=$(echo "${private_key}" | openssl rsa -pubout 2>/dev/null )

  ${KUSTOMIZE} edit add secret traefik-jwt-keys \
    --disableNameSuffixHash \
    --from-literal=public="${public_key}" \
    --from-literal=private="${private_key}"

  # update traefik config
  ${YQ} eval \
    '.http.middlewares.jwt-authorizer.plugin.jwt.keys[0]="'"${public_key}"'"' \
    --inplace "${f}/config/server/proxy-config/dynamic/config.yaml"
fi

# updating config locally
${KUSTOMIZE} edit set namespace "$1"
${KUSTOMIZE} edit add configmap rest-api-server-config \
        --behavior=replace \
        --from-literal=log.level="${SERVER_LOG_LEVEL}" \
        --from-literal=kubesaw.namespace="${TOOLCHAIN_HOST}"

cd "${f}/config/server"
${KUSTOMIZE} edit set image workspaces/rest-api="$2"

if [[ -n "${MANIFEST_TARBALL}" ]]; then
    # save config to a tarball
    cd "${f}"
    tar -caf "${MANIFEST_TARBALL}" config/
else
    # apply config
    cd "${f}/config/default"
    ${KUSTOMIZE} build . | ${KUBECLI} apply -f -
fi

# cleanup
rm -r "${f}"
