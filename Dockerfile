# Build Stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the bot application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Build the migrate tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bot .
COPY --from=builder /build/migrate .

# Copy Data directory with CSV files
COPY Data ./Data

# Create directory for database
RUN mkdir -p /app/data

# Create directory for config
RUN mkdir -p /app/config

VOLUME ["/app/data"]

CMD ["./bot"]
