# Стадия сборки
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Стадия запуска
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .

# Создаем непривилегированного пользователя для безопасности
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080
CMD ["./main"]