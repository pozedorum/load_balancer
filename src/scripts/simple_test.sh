#!/bin/bash

# Базовый порт сервера
BASE_PORT=8080

# Количество серверов
NUM_SERVERS=3

# Функция для отправки запроса на сервер
send_request() {
  local port=$1
  local server="http://localhost:$port"
  local response=$(curl -s -w "%{http_code}" $server/health)
  echo "Ответ от сервера $server: $response"
}

# Проверяем соединение с балансировщиком нагрузки
send_request $BASE_PORT

# Проверяем соединение с каждым сервером
for ((i=1; i<=NUM_SERVERS; i++)); do
  port=$((BASE_PORT + i))
  send_request $port
done