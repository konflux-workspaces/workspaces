# REST API Server

The [REST API Server][server-folder] is respecting the [hexagonal architecture][hexagonal-architectrue] architectural pattern and [Command Query Responsibility Segregation (CQRS) pattern][cqrs-pattern].

* the [`api` folder][server-api-folder] contains the Go code for Operator's CRDs
* the [`config` folder][server-config-folder] contains the YAML manifests
* the [`rest` folder][server-rest-folder] contains the REST over HTTP layer.
    This package main responsibilities are to map HTTP requests to `core`'s Commands and Queries, trigger the `core` logic, and map `core`'s Responses back to HTTP Responses.
* the [`core` folder][server-core-folder] contains the application main logic.
    It validates requests, access the persistence layer to fetch the correct data and produces a response. 
* the [`persistence` folder][server-persistence-folder] contains the persistence layer code.
    More in details, it contains caches and Kubernetes client implementation.

Following the flow of an HTTP Request, the request will be initially processed by the code in the [`rest` package][server-rest-folder].
It validates the HTTP Request and builds a Command or Query to use for invoking the handlers in the [`core` package][server-core-folder].
The `core` package performs validation, authorization, may apply some transformation, before invoking the logic in the [`persistence` package][server-persistence-folder].
In case of a Command it will perform update or create, in case of a Query it will retrieve some data.
Finally, the `core` will build a Response and provide it back to the `rest` package which will map it to an HTTP Response.

## Run Tests

To run unit tests you can execute the `make test` command.

To run e2e tests, take a look at the [Run End-to-End Test](./e2e/run-tests.md) section.

<!-- external links -->

[server-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server
[server-api-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server/api
[server-config-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server/config
[server-rest-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server/rest
[server-core-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server/core
[server-persistence-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/server/persistence

[hexagonal-architectrue]: https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)
[cqrs-pattern]: https://martinfowler.com/bliki/CQRS.html
