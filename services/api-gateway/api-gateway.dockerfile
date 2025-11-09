ARG binary=apiGateway
ARG service=api-gateway

FROM golang:1.25-alpine AS builder

ARG binary
ARG service

WORKDIR /app

# Copy proto and common modules first (dependencies)
COPY proto /app/proto
COPY common /app/common

COPY $service /app/$service

WORKDIR /app/${service}

RUN go mod download

# Build the application
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o $binary ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

ARG binary
ARG service

COPY --from=builder /app/$service/$binary .

EXPOSE 80

CMD ["./${binary}"]
