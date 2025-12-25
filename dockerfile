# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install git and certificates (needed for private modules or HTTPS)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Leverage Docker cache by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with optimizations:
# -ldflags="-s -w" removes symbol tables and debug info to shrink size
# CGO_ENABLED=0 ensures a static binary for portability
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./main.go

# Stage 2: Runtime
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your app runs on
EXPOSE 8080

# Run the binary
CMD ["./main"]