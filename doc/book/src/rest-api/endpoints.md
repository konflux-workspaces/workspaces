# Endpoints

## Workspaces

This section details the endpoints for [Workspaces](./crds.md) exposed by the REST API Server.


### `/apis/workspaces.konflux.io/v1alpha1/`

Requests to this workspace will always be authorized, the result varies with respect to the access the requesting user has.


#### `GET`

This endpoint returns the list of all the workspaces the user has access to.
The workspace can be own by different user.


### `/apis/workspaces.konflux.io/v1alpha1/namespaces/{owner}/workspaces/{workspace}`

Requests to this workspace will be authorized only if the user has access to the workspace `{workspace}` owned by the user `{owner}`.


#### `GET`

Returns the details for the workspace `{workspace}` owned by the user `{owner}`.


#### `PUT`

> Only the owner is allowed to perform this operation.

Allows the user to update the `spec` of the workspace `{workspace}` owned by the user `{owner}`.
