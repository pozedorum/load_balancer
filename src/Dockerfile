# Используем базовый образ Go
FROM golang:1.21-alpine

# Устанавливаем необходимые утилиты:
# - git и make для сборки
# - ca-certificates для HTTPS
# - jq для работы с JSON
RUN apk add --no-cache git make ca-certificates jq

# Создаем рабочую директорию
WORKDIR /app

# Копируем все файлы проекта
COPY . .

# Скачиваем зависимости
RUN go mod download


# Открываем порты
EXPOSE 8080 8081 8082 8083 8084 8085
