#!/bin/bash

set -ex

# parse input
BRANCH=${BRANCH:-pubviewer-mvp}

# checkout
git clone --depth 2 --branch "${BRANCH}" https://github.com/filariow/host-operator.git
git clone --depth 1 --branch "${BRANCH}" https://github.com/filariow/registration-service
