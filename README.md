# WhoDidThat Controller

A Kubernetes admission controller that automatically tracks resource creation and modification by adding user information labels.

## This project is in early development(could be used in production).
Things to do:
- Add tests
- Add documentation
- Add examples
- Add features:
    - Other labels
    - Think about annotations
    - Metrics


## Overview

WhoDidThat Controller is a Kubernetes mutating admission webhook that adds audit labels to resources, tracking who created and last modified them. It adds the following labels:
- `audit.k8s.io/created-by`: The user who created the resource
- `audit.k8s.io/last-modified-by`: The user who last modified the resource

## Features

- Tracks resource creation and modification events
- Adds user information as Kubernetes labels
- Supports all Kubernetes resource types
- Multi-architecture support (AMD64 and ARM64)
- Minimal footprint using distroless container image

## Prerequisites

- Kubernetes cluster 1.16+
- kubectl configured to access your cluster
- cert-manager (recommended for TLS certificate management)

## Installation

### Install the chart

```bash
helm install whodidthat oci://ghcr.io/x0ddf/whodidthat
```

### Configuration

The following table lists the configurable parameters of the WhoDidThat chart and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of controller replicas | `1` |
| `image.repository` | Image repository | `ghcr.io/x0ddf/whodidthat-controller` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `certManager.enabled` | Enable cert-manager integration | `true` |
| `webhook.failurePolicy` | Webhook failure policy | `Ignore` |
| `webhook.timeoutSeconds` | Webhook timeout | `5` |

## Usage

Once installed, the controller automatically adds labels to resources when they are created or modified.

Example resource with added labels:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    audit.k8s.io/created-by: john@example.com
    audit.k8s.io/last-modified-by: jane@example.com
spec:
  containers:
  - name: nginx
    image: nginx:latest
```

## Development

### Prerequisites

- Go 1.23+
- Docker

### Building

Build the binary:
```bash
go build -o whodidthat-controller
```

Build multi-arch container image:
```bash
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag ghcr.io/<your-github-username>/whodidthat-controller:latest \
    --push \
    .
```

### Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


