# `REST API`

The REST API Server exposes Kubernetes Server compatible [endpoints](./endpoints.md) for [Workspaces](./crds.md).

Under the hoods, the REST API Server works with [InternalWorkspaces](../operator/crds.md).
[Workspaces](./crds.md) are never persisted on storage.
Any allowed change performed on Workspaces is reflected by the REST API Server on [InternalWorkspaces](../operator/crds.md).

Hence, one of the REST API Server's main aims is to provide its users a Kubernetes-like experience on Workspace *virtual* custom resources.

Another responsibility of the REST API Server is to authenticate the users performing requests and provide them only the data they're allowed to have.
The [Authorization](./auth.md) logic is simple at the moment.
Users can only access the workspace they own and the ones that has been shared with them.

