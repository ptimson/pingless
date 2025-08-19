# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary
RUN go build -o pingless main.go

# Runtime stage
FROM alpine:latest

# Install sshpass
RUN apk add --no-cache sshpass openssh

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/pingless ./

# Install ping utility for ICMP (required by go-ping)
RUN apk add --no-cache iputils

# Run the binary
CMD ["./pingless"]
