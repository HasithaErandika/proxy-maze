FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy-maze .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/proxy-maze .

EXPOSE 8080

CMD ["./proxy-maze"]
