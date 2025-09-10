# ---------- Stage 1: Build ----------
FROM golang:1.23-bullseye AS builder

WORKDIR /app

# Устанавливаем зависимости для confluent-kafka-go и goose
RUN apt-get update && apt-get install -y \
    gcc g++ librdkafka-dev pkg-config \
    git postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Установка goose (используем go install вместо apk)
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем весь проект
COPY . .

# Подтягиваем зависимости и собираем
RUN go mod tidy && go mod download
RUN go build -o main ./cmd/

# Копируем миграции
COPY migrations ./migrations

# ---------- Stage 2: Runtime ----------
FROM debian:bullseye-slim

WORKDIR /app

# Устанавливаем только рантайм библиотеки
RUN apt-get update && apt-get install -y \
    librdkafka1 ca-certificates \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Копируем бинарь, конфиг, статику и goose
COPY --from=builder /app/main /app/main
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations /app/migrations
COPY config.yaml /app/config.yaml
COPY --from=builder /app/static /app/static

# Указываем путь к конфигу через ENV
ENV CONFIG_PATH=/app/config.yaml

# Пробрасываем порт
EXPOSE 8081

ENV DB_URL=postgres://user:password@postgres:5432/orders_db?sslmode=disable

# Entrypoint скрипт
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
