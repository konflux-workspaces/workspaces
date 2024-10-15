# `Custom Resource Definitions (CRDs)`

InternalWorkspaces contains information required to build [Workspaces](../rest-api/crds.md) and to manage related [KubeSaw](https://github.com/codeready-toolchain)'s resources.

```yaml
apiVersion: workspaces.konflux-ci.dev/v1alpha1
kind: InternalWorkspace
metadata:
    namespace: workspaces-system
    name: my-workspace-7ghf2
spec:
    displayName: my-workspace
    visibility: community | private
    owner:
        jwtInfo:
            email: string
            sub: string
            userId: string
status:
    space:
        # whether it is the home KubeSaw's Space for the user or not
        isHome: true | false
        # the name of the related KubeSaw's Space
        name: my-workspace-7ghf2
    owner:
        # the name of the owner's KubeSaw's UserSignup
        username: string
    conditions:
        type: string
        status: True | False | Unknown
        reason: string
        message: string
        lastTransitionTime: time
```
