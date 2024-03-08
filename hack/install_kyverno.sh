#!/bin/bash

KYVERNO_VERSION=${KYVERNO_VERSION:-v1.10.0}

kubectl create -f "https://github.com/kyverno/kyverno/releases/download/$KYVERNO_VERSION/install.yaml" -n kyverno
