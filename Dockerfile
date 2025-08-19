FROM golang:alpine AS builder

WORKDIR /usr/src/app

# Install dependencies
RUN apk add --no-cache git tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .
RUN go build -v -o /usr/local/bin/bazarr-sync ./cmd/bazarr-sync/

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata python3

# Copy binary from builder
COPY --from=builder /usr/local/bin/bazarr-sync /usr/local/bin/bazarr-sync

# Create symbolic link for shorter command
RUN ln -s /usr/local/bin/bazarr-sync /usr/local/bin/bs

# Create default cache files
RUN touch /movies-cache /shows-cache

# Create config directory
WORKDIR /config

CMD ["bazarr-sync"]