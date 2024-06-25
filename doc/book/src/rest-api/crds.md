# `Custom Resource Definitions (CRDs)`

The Workspace Custom Resource Definition is simple and minimal.
It just contains the information required by the Konflux UI.

Workspaces are never persisted on storage, but always calculated from [InternalWorkspaces](../operator/crds.md).
Any allowed change performed on Workspaces is reflected by the REST API Server on [InternalWorkspaces](../operator/crds.md).

```yaml
apiVersion: workspaces.konflux.io
kind: Workspaces
metadata:
    namespace: owner-name
    name: my-workspace
spec:
    visibility: community | private
status:
    owner:
        email: string
    space:
        name: string
    conditions:
        type: string
        status: True | False | Unknown
        reason: string
        message: string
        lastTransitionTime: time
```
