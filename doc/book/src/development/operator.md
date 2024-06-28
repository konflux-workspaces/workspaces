# Operator

The [operator code][operator-folder] is using the [Operator-SDK][operator-sdk].

* the [`api` folder][operator-api-folder] contains the Go code for Operator's CRDs
* the [`config` folder][operator-config-folder] contains the YAML manifests
* the [`internal` folder][operator-internal-folder] contains the code for the reconcilers 

## Run Tests

To run Unit tests you can execute the `make test` command.

To run e2e tests, take a look at the [Run End-to-End Test](./e2e/run-tests.md) section.


<!-- external links -->

[operator-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/operator
[operator-api-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/operator/api
[operator-config-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/operator/config
[operator-internal-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/operator/internal

[operator-sdk]: https://sdk.operatorframework.io
