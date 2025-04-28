#!/bin/bash

# Скрипт для проверки доступности серверов
TIMEOUT=1  # Таймаут в секундах
ATTEMPTS=15 # Количество попыток
BALANCER_URL="http://localhost:8080"
BACKEND_URLS=("http://localhost:8081" "http://localhost:8082" "http://localhost:8083")



# Функция для проверки URL
check_url() {
  local url=$1
  local success=0
  local total=0

  echo "Проверяем $url..."
  for ((i=1; i<=$ATTEMPTS; i++)); do
    if curl -s --max-time $TIMEOUT "$url/health" > /dev/null; then
      echo "Попытка $i: УСПЕХ"
      ((success++))
    else
      echo "Попытка $i: ОШИБКА (таймаут ${TIMEOUT}с)"
    fi
    ((total++))
    sleep 0.5 # Небольшая пауза между проверками
  done

  echo "Итого по $url: $success успешных из $total попыток"
  echo "----------------------------------------"
}


# Проверяем backend-серверы
for url in "${BACKEND_URLS[@]}"; do
  check_url "$url"
done

echo "Проверка завершена"