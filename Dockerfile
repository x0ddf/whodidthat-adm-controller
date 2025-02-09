# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for multi-arch support
ARG TARGETARCH
ARG TARGETOS

# Build the application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s" \
    -o /whodidthat-controller

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /whodidthat-controller .

# Expose the port
EXPOSE 8443

# Container runs as nonroot (uid:65532, gid:65532) by default in distroless
ENTRYPOINT ["/whodidthat-controller"] 