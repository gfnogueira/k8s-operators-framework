# ============================================================================
# Multi-stage Dockerfile for the MyApp Operator
# Stage 1: Build the Go binary
# Stage 2: Minimal runtime image
# ============================================================================

# Build stage
FROM golang:1.22 AS builder

WORKDIR /workspace

# Cache dependencies
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build — CGO disabled for static binary, targeting linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Runtime stage — distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
