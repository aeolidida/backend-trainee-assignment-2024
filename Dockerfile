FROM --platform=linux/amd64 golang:1.22.1-alpine3.19 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o main cmd/main.go

FROM --platform=linux/amd64 alpine:3.19

WORKDIR /app

COPY --from=builder /build/main /app/main
COPY --from=builder /build/config/. /app/config/.

RUN chmod +x /app/main

ENV CONFIG_PATH=${CONFIG_PATH} \
    DB_USER=${DB_USER} \
    DB_PASSWORD=${DB_PASSWORD} \
    AUTH_SECRET_KEY=${AUTH_SECRET_KEY} \
    REDIS_PASSWORD=${REDIS_PASSWORD} \
    RABBITMQ_DEFAULT_USER=${RABBITMQ_DEFAULT_USER} \
    RABBITMQ_DEFAULT_PASS=${RABBITMQ_DEFAULT_PASS}

EXPOSE 8080

CMD ["./main"]