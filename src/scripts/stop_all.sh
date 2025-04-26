#!/bin/bash

# Останавливаем все процессы
if [ -f pids.txt ]; then
  echo "Останавливаем запущенные процессы..."
  while read pid; do
    if ps -p $pid > /dev/null; then
      kill $pid
      echo "Остановлен процесс $pid"
    fi
  done < pids.txt
  rm pids.txt
fi

# Дополнительная очистка
pkill -f "cmd/backend/main.go"
pkill -f "cmd/balancer/main.go"  # <- Изменение здесь

echo "Все сервисы остановлены"