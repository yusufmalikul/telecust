# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telecust .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/telecust .

# Copy web directory for frontend UI
COPY web ./web

# Create directory for database
RUN mkdir -p /data

# Set database path for Docker environment
ENV DB_PATH=/data/telecust.db

# Expose port (adjust if your API uses a different port)
EXPOSE 8080

# Run the application
CMD ["./telecust"]
