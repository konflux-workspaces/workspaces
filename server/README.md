# Workspaces API Server

This folder contains the Workspaces API Server.

This server implements the REST over HTTP endpoints users can invoke to retrieve information about:

* All the workspaces they have visibility on
* Details of a given workspace they have visibility on

## Repo structure

The Server code is based on Hexagonal architecture and CQRS.

* Business Logic is stored under `core`
* Driving Adapters
    * REST over HTTP server implementation is stored under `rest`
* Driven Adapters
    * Read-Model Cache under `persistence/cache`
    * Write-Model Kubernetes client under `persistence/kube`

