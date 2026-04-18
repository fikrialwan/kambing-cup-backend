# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies (build-base is required for CGO)
RUN apk add --no-cache git ca-certificates build-base

WORKDIR /app

# 1. Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source code and migrations
COPY . .

# 3. Build static binary with CGO enabled
# We keep CGO_ENABLED=1 because goheif requires it
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o main ./main.go

# Stage 2: Runtime (Production Image)
FROM alpine:3.19  

# Security: Install CA certs and tzdata
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 4. Copy the compiled binary
COPY --from=builder /app/main .

# 5. IMPORTANT: Copy migrations folder for the app to run on startup
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/serviceAccountKey.json .

# Use a non-root user for better security in production
# RUN adduser -D appuser
# USER appuser

EXPOSE 8080

# Run the binary
CMD ["./main"]