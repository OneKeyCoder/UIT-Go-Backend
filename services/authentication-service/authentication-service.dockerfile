ARG binary=authService
ARG service=authentication-service

FROM golang:1.25-alpine AS builder

ARG binary
ARG service

WORKDIR /app

# Copy proto and common modules first (dependencies)
COPY proto /app/proto
COPY common /app/common

# Get build dependencies first
COPY $service/go.mod $service/go.sum /app/$service/

WORKDIR /app/${service}

RUN go mod download

# Copy the rest of the service
COPY $service /app/$service

RUN CGO_ENABLED=0 go build -o $binary ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

ARG binary
ARG service

COPY --from=builder /app/$service/$binary .

EXPOSE 80
EXPOSE 50051

ENV entrypoint=$binary

ENTRYPOINT ./${entrypoint}
