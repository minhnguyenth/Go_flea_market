# ===== Build Stage =====
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
# Since the build context in docker-compose.yaml is the project root,
# specify the path from the context (project root) rather than
# the relative path from the Dockerfile location (docker/app).
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code (copy entire project root)
COPY . .

# Build application (specify main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./main.go

# ===== Runtime Stage =====
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy executable from build stage
COPY --from=builder /app/main .

# (Optional) Copy configuration files and templates required by application
# COPY app/config.yaml .
# COPY app/templates ./templates
# COPY app/static ./static

# Port that application listens on (match docker-compose.yaml)
EXPOSE 8081

# Run application when container starts
CMD ["./main"] 