# WhoDidThat Controller

A Kubernetes admission controller that automatically tracks resource creation and modification by adding user information labels.

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

### Using Helm (Recommended)

1. Add the repository:
   ```bash
   helm repo add whodidthat https://x0ddf.github.io/whodidthat-controller
   ```

