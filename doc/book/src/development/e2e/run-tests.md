# Run End-to-End tests

To run the End-to-End tests, you need a QUAY repository and admin access to an OpenShift cluster where [KubeSaw][kubesaw] and [Konflux-Workspaces][konflux-workspaces] are running.

## All in one script

To easily setup the cluster you can refer to the [`./hack/demo.sh` script][hack-demo-script].
This script will install [KubeSaw][kubesaw] and [Konflux-Workspaces][konflux-workspaces], and then execute e2e tests.

## Step by step

### Install dependencies

You need to define some variables to use in the next steps.

```sh
# the tag to use for KubeSaw images build at step 1
export IMAGE_TAG=e2e-test

# the quay.io namespace to use in the next steps
export QUAY_NAMESPACE=my-quay-namespace
```

> Tip
>
> By default the scripts will use `docker`.
> If you want to use a different tool for building and pushing containers, you can export the `IMAGE_BUILDER` variable.
> As an example, to use podman you will `export IMAGE_BUILDER=podman`.

#### 1. Build KubeSaw fork

As first thing, you'll need to build and push the KubeSaw fork from Konflux-Workspaces.
> The [`ci/toolchain_manager.sh` script][toolchain-manager-script] provides help to complete this step.

```sh
./ci/toolchain_manager.sh publish "$IMAGE_TAG" -n "$QUAY_NAMESPACE"
```

#### 2. Install KubeSaw

Once images from our KubeSaw fork are built and published, you need to deploy them in the cluster.
> The [`ci/toolchain_manager.sh` script][toolchain-manager-script] provides help to complete this step.

```sh
./ci/toolchain_manager.sh deploy "$IMAGE_TAG" -n "$QUAY_NAMESPACE"
```

#### 3. Install Konflux-Workspace

To build and install the Konflux-Workspaces, you can use the [`hack/workspaces_install.sh` script][workspaces-install-script].

```sh
# remember to export QUAY_NAMESPACE=my-quay-namespace
./hack/workspaces_install.sh
```

### Run the tests

Now that the dependencies are installed, you can run the End-to-End tests by executing the following command:

```sh
make -C e2e test
```

<!-- external links -->

[kubesaw]: https://github.com/codeready-toolchain/
[konflux-workspaces]: https://github.com/konflux-workspaces/workspaces

[hack-demo-script]:  https://github.com/konflux-workspaces/workspaces/blob/main/hack/demo.sh
[toolchain-manager-script]: https://github.com/konflux-workspaces/workspaces/blob/main/ci/toolchain_manager.sh
[workspaces-install-script]: https://github.com/konflux-workspaces/workspaces/blob/main/hack/workspaces_install.sh 
