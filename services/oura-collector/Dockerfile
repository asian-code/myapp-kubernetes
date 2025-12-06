# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.work and modules
COPY go.work ./
COPY shared ./shared
COPY oura-collector ./oura-collector

WORKDIR /app/oura-collector

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/oura-collector/main .

# Run as non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root

USER appuser

CMD ["./main"]
