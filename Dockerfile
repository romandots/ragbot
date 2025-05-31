# Builder stage
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /app

# Copy source and go modules
COPY go.mod go.sum ./
RUN go mod download
COPY ./ .

# Build Go binary
RUN go build -o ragbot main.go

# Final image
FROM alpine:latest
RUN apk add --no-cache libstdc++ libgcc
WORKDIR /app
COPY --from=builder /app/ragbot ./ragbot

EXPOSE 8080
ENTRYPOINT ["/app/ragbot"]