#!/bin/bash

CONFIG_FILE="config/servers.json"
LOG_DIR="logs"
PID_FILE="$LOG_DIR/server.pids"

# Функция для отправки запроса на сервер
send_request() {
  local port=$1
  local server="http://localhost:$port"
  local response=$(curl -s -w "%{http_code}" $server/health)
  echo "Ответ от сервера $server: $response"
}

# Проверяем соединение с балансировщиком нагрузки
send_request 8080

# Проверяем соединение с каждым сервером
for server in $(jq -c '.[]' $CONFIG_FILE); do
  port=$(jq -r '.port' <<< "$server")
  send_request $port
done