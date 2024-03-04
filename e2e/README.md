# End-To-End tests

End-to-End tests implemented here leverages [Gherkin](https://cucumber.io/docs/gherkin/) and the [godog](https://github.com/cucumber/godog) framework.

Features are stored under [features](./features).

## Run tests

Prerequisites:

* An OpenShift cluster exists
* Current shell is logged-in the OpenShift cluster as `admin` (e.g. `oc login`)
* [KubeSaw](https://github.com/codeready-toolchain/) is deployed and configured as a multi-cluster environment
* [Workspaces Operator](../operator/) is deployed  
* [Workspaces REST API Server](../server/) is deployed

To execute all the tests

```bash
make test
```

## Development

To test a single test scenario, you can add the tag `@wip` before the scenario definition, then you can run the single test with the following command:

```bash
make wip
```
