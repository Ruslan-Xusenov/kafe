# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary libraries (like ca-certificates for HTTPS)
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/main .
COPY .env.example .env

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
