FROM golang:alpine AS builder

WORKDIR /usr/src/app

# Install dependencies
RUN apk add --no-cache tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .
RUN go build -v -o /usr/local/bin/bazarr-sync ./cmd/bazarr-sync/

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /usr/local/bin/bazarr-sync /usr/local/bin/bazarr-sync

# Create default cache files
RUN touch /movies-cache /shows-cache

# Create config directory
WORKDIR /config

CMD ["bazarr-sync"]