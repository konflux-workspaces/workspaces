#!/bin/bash

set -e -o pipefail

KYVERNO_VERSION="${KYVERNO_VERSION:-v1.12.2}"

kubectl create \
  -f "https://github.com/kyverno/kyverno/releases/download/${KYVERNO_VERSION}/install.yaml" \
  -n kyverno
