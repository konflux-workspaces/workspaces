#!/bin/bash

# Builds and publishes kubesaw components to a quay repository.  Used as a part
# of the CI process for testing in openshift-ci.

set -e

# parse input
QUAY_NAMESPACE=${QUAY_NAMESPACE:-konflux-workspaces}
TAG=${2:-e2etest}
KUBECLI=${KUBECLI:-kubectl}

HOST_OPERATOR_REPO=${HOST_OPERATOR_REPO:-https://github.com/filariow/host-operator.git}
MEMBER_OPERATOR_REPO=${MEMBER_OPERATOR_REPO:-https://github.com/codeready-toolchain/member-operator.git}
TOOLCHAIN_E2E_REPO=${TOOLCHAIN_E2E_REPO:-https://github.com/sadlerap/toolchain-e2e.git}
REGISTRATION_SERVICE_REPO=${REGISTRATION_SERVICE_REPO:-https://github.com/filariow/registration-service}

HOST_OPERATOR_BRANCH=${HOST_OPERATOR_BRANCH:-${BRANCH:-pubviewer-mvp}}
MEMBER_OPERATOR_BRANCH=${MEMBER_OPERATOR_BRANCH:-${BRANCH:-master}}
TOOLCHAIN_E2E_BRANCH=${TOOLCHAIN_E2E_BRANCH:-${BRANCH:-pubviewer-mvp}}
REGISTRATION_SERVICE_BRANCH=${REGISTRATION_SERVICE_BRANCH:-${BRANCH:-pubviewer-mvp}}

function clone {
    git clone --depth 2 --branch "${2}" "${1}"
}

# build & publish toolchain components
function publish {
    clone "${HOST_OPERATOR_REPO}"        "${HOST_OPERATOR_BRANCH}"
    clone "${MEMBER_OPERATOR_REPO}"      "${MEMBER_OPERATOR_BRANCH}"
    clone "${TOOLCHAIN_E2E_REPO}"        "${TOOLCHAIN_E2E_BRANCH}"
    clone "${REGISTRATION_SERVICE_REPO}" "${REGISTRATION_SERVICE_BRANCH}"

    make -C toolchain-e2e publish-current-bundles-for-e2e \
        FORCED_TAG="${TAG}" \
        QUAY_NAMESPACE="${QUAY_NAMESPACE}" \
        REG_REPO_PATH=../registration-service \
        HOST_REPO_PATH=../host-operator \
        MEMBER_REPO_PATH=../member-operator \
        CI=true
}

function deploy {
    clone "${TOOLCHAIN_E2E_REPO}" "${BRANCH}"

    make -C toolchain-e2e deploy-published-operators-e2e \
        FORCED_TAG="${TAG}" \
        QUAY_NAMESPACE="${QUAY_NAMESPACE}" \
        SECOND_MEMBER_MODE=false \
        CI=true \
        HOST_NS=toolchain-host-operator \
        MEMBER_NS=toolchain-member-operator

    ${KUBECLI} patch \
        toolchainconfigs.toolchain.dev.openshift.com config \
        -n toolchain-host-operator \
        --patch='{"spec":{"publicViewer":{"username":"public-viewer"}}}' \
        --type=merge

    ${KUBECLI} delete pods --all --namespace toolchain-host-operator
}

function usage {
    set -o pipefail
    echo "${0} usage:"
    grep -G "\s\+[a-z]\+)\s\+#" "$0" | sed -e 's/^\s\+\([a-z]\+\))\s\+#\s\+\(\(.*\)\s\+##\s\+\)\?\(.*\)$/\t\1\t\3\n\t\t\4/g'
    exit 0
}

function use_tmp_dir {
    # create a temporary direction
    f=$(mktemp --directory /tmp/toolchain.XXXX)
    cd "${f}"
}

function set_namespace {
    # TODO(sadlerap): convert this to using something like getopt if we need to
    # support more arguments
    if [[ -n "$1" ]]; then
        if [[ -n "$2" ]]; then
            echo "Using namespace ${2}"
            export QUAY_NAMESPACE=${2}
        else
            echo "ERROR: expected namespace in arguments, but none was found"
        fi
    fi
}

case "${1}" in
    publish) # IMAGE_TAG [-n QUAY_NAMESPACE] ## build and publish images to a registry
        set_namespace "${3}" "${4}"
        use_tmp_dir
        publish
        ;;
    deploy)  # IMAGE_TAG [-n QUAY_NAMESPACE] ## deploy images to a live cluster
        set_namespace "${3}" "${4}"
        use_tmp_dir
        deploy
        ;;
    help)    # display this help message
        usage
        ;;
    *)
        usage
        ;;
esac
