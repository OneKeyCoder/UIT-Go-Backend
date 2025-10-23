FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy proto and common modules first (dependencies)
COPY proto /app/proto
COPY common /app/common

# Copy location-service
COPY location-service /app/location-service

# Set working directory to location-service
WORKDIR /app/location-service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o locationServiceApp ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/location-service/locationServiceApp .

EXPOSE 80

CMD ["./locationServiceApp"]
