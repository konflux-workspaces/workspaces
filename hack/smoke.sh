#!/bin/bash
#
# Usage:
#
# To use this script for smoke testing, you'll need to extract a valid user
# token from the relevant URL.  Note that tokens aren't portable, i.e. a token
# from console.redhat.com won't work on console.dev.redhat.com and vice versa.

function check_command {
    OUTPUT="$(command -v "${1}")"
    if [[ "${OUTPUT}" = "" ]]; then
        echo "${1} required for smoke tests!"
        exit 1
    fi
}

function display_help {
    echo "help for $(basename "${0}")"
    echo "-e    --environment   Sets the environment to run smoke tests on"
    echo "-h    --help          Print this help message"
    echo "-t    --token         Specifies the JWT token to use"
}

function match_environment {
    case "${1}" in
        staging-internal)
            echo "Internal staging is not yet supported!"
            exit 1
            ;;
        production-internal)
            echo "Internal production is not yet supported!"
            exit 1
            ;;
        staging-external)
            CLUSTER_URL="https://console.dev.redhat.com"
            ;;
        production-external)
            CLUSTER_URL="https://console.redhat.com"
            ;;
        *)
            CLUSTER_URL="${1}"
    esac
}

while [[ $# -gt 0 ]]; do
    case "${1}" in
        -h | --help)
            display_help
            exit 0
            ;;
        -t | --token)
            shift
            TOKEN="${1}"
            shift
            ;;
        -e | --environment)
            shift
            match_environment "${1}"
            shift
            ;;
        -u | --user)
            shift
            USERNAME="${1}"
            shift
            ;;
        *)
            echo "Unrecognized command ${1}!"
            display_help
            exit 1
    esac
done

if [[ -z "${CLUSTER_URL}" || -z "${TOKEN}" || -z "${USERNAME}" ]]; then
    if [[ -z "${CLUSTER_URL}" ]]; then
        echo "Expected environment, found none"
    fi
    if [[ -z "${TOKEN}" ]]; then
        echo "Expected token, found none"
    fi
    if [[ -z "${USERNAME}" ]]; then
        echo "Expected username, found none"
    fi

    display_help
    exit 1
fi

check_command curl
check_command jq

RESPONSE=$(curl --oauth2-bearer "${TOKEN}" -sSfL "${CLUSTER_URL}/api/k8s/apis/workspaces.konflux-ci.dev/v1alpha1/workspaces")
ARGS=".items[] | select(.metadata.namespace == \"${USERNAME}\")"
OUTPUT="$(echo "${RESPONSE}" | jq "${ARGS}")"
if [[ "${OUTPUT}" = "" ]]; then
    LOG_FILE="${TMPDIR:-/tmp}/workspaces.$(date +%s)"

    echo "Failed to find default namespace for user \"${USERNAME}\"!"
    echo "${RESPONSE}" > "${LOG_FILE}"
    echo "Response output saved in ${RESPONSE}"

    exit 1
fi

echo "Smoke tests succeeded!"
