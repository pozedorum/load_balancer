#!/bin/bash

# Создаем папку для логов
mkdir -p logs

# Переходим в корень проекта
cd $(dirname "$0")/..

# Запускаем 3 тестовых сервера
for port in {8081..8083}; do
  go run cmd/backend/main.go $port > "logs/backend_$port.log" 2>&1 &
  echo $! >> pids.txt
  echo "Запущен backend-сервер на порту $port (PID: $!)"
done

# Запускаем балансировщик нагрузки
go run cmd/balancer/main.go > "logs/balancer.log" 2>&1 &  # <- Изменение здесь
echo $! >> pids.txt
echo "Запущен балансировщик нагрузки (PID: $!)"

echo "----------------------------------------"
echo "Система готова к работе!"
echo "Для теста выполните: curl http://localhost:8080"
echo "Для остановки нажмите Ctrl+C"
echo "----------------------------------------"

wait