FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrator .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/migrator .

# Copy migrations
COPY ../migrations /migrations

ENV MIGRATIONS_PATH=file:///migrations

ENTRYPOINT ["./migrator"]
