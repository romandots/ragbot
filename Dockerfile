# Builder stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /app

# Copy source and go modules
COPY go.mod go.sum ./
RUN go mod download -x
COPY ./ .

# Build Go binary
RUN go build -o ragbot cmd/ragbot/main.go

# Final image
FROM alpine:latest
RUN apk add --no-cache libstdc++ libgcc
WORKDIR /app
COPY --from=builder /app/ragbot ./ragbot

EXPOSE 8080
ENTRYPOINT ["/app/ragbot"]