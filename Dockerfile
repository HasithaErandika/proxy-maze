FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o proxy-maze .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/proxy-maze .

EXPOSE 8080

CMD ["./proxy-maze"]
