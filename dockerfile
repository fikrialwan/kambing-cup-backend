# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# 1. Cache dependencies (speeds up GitHub Action builds)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source code and migrations
COPY . .

# 3. Build optimized static binary
# -ldflags="-s -w" reduces binary size by ~25%
# CGO_ENABLED=0 ensures the binary runs on the slim Alpine image
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./main.go

# Stage 2: Runtime (Production Image)
FROM alpine:3.19  

# Security: Install CA certs for HTTPS/Postgres connections
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 4. Copy the compiled binary
COPY --from=builder /app/main .

# 5. IMPORTANT: Copy migrations folder for the app to run on startup
COPY --from=builder /app/migrations ./migrations

# Use a non-root user for better security in production
# RUN adduser -D appuser
# USER appuser

EXPOSE 8080

# Run the binary
CMD ["./main"]