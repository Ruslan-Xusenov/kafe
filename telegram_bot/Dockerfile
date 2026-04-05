# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the bot binary
RUN CGO_ENABLED=0 GOOS=linux go build -o kafe-bot .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary libraries
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/kafe-bot .

# Command to run
CMD ["./kafe-bot"]
