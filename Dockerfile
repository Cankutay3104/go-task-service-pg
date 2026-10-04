# Stage 1: The Builder
FROM golang:1.27-alpine AS builder

RUN apk add --no-cache ca-certificates

# Create and enter the virtual /app directory inside the container
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy your local source code into the container
COPY . .

# Compile the static binary directly into /app/task-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o task-service cmd/api/main.go

# Stage 2: The Runtime
FROM scratch

# Safely copy certificates into a directory using the trailing slash
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Retrieve the standalone binary from the builder
COPY --from=builder /app/task-service /task-service

EXPOSE 8080

ENTRYPOINT ["/task-service"]