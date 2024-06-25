# Auth

The REST API Server authenticates and authorizes requests before processing them.


## Authentication

The Authentication is performed by a Traefik sidecar configured for validate the request JWT token.
The sidecar also extracts meaningful fields and injects them as Headers before proxying the request to the REST API Server.

Hence, configuring Authentication is as easy as correctly configuring the Traefik sidecar to use the correct key to validate the JWTs.


## Authorization

For authorizing requests, the REST API server fetches information from [KubeSaw](https://github.com/codeready-toolchain)'s resources.
Namely, UserSignup and SpaceBindings are checked.

To fetch the correct resources, the REST API Server matches the JWT's `sub` and UserSignup's `spec.sub` fields.
