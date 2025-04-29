#!/bin/bash

# Параметры теста
TARGET_URL="http://localhost:8080"
REQUESTS_PER_IP=40
DELAY=0.05

# Массив IP для тестирования
IPS=("192.168.1.1" "10.0.0.5" "172.16.0.10" "1.2.3.4")

# Инициализация счетчиков
init_counters() {
    for ip in "${IPS[@]}"; do
        eval "success_${ip//./_}=0"
        eval "limit_${ip//./_}=0"
        eval "error_${ip//./_}=0"
    done
}

# Получение значений счетчиков
get_counter() {
    local ip=$1
    local type=$2
    eval "echo \$${type}_${ip//./_}"
}

# Увеличение счетчика
inc_counter() {
    local ip=$1
    local type=$2
    eval "(( ${type}_${ip//./_}++ ))"
}

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Функция для отправки запроса
send_request() {
    local ip=$1
    local exec_time=$(( RANDOM % 50 + 50 ))

    local response=$(curl -s -w "\n%{http_code}" -H "X-Forwarded-For: $ip" -H "Execution-Time: $exec_time" "$TARGET_URL")
    local status=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    case $status in
        202)
            inc_counter "$ip" "success"
            printf "${GREEN}Request from %s: SUCCESS (%dms)${NC}\n" "$ip" "$exec_time"
            ;;
        429)
            inc_counter "$ip" "limit"
            printf "${YELLOW}Request from %s: RATE LIMITED (%dms)${NC}\n" "$ip" "$exec_time"
            ;;
        *)
            inc_counter "$ip" "error"
            printf "${RED}Request from %s: ERROR %s (%dms)${NC}\n" "$ip" "$status" "$exec_time"
            ;;
    esac
}

# Инициализация
init_counters

# Заголовок теста
printf "\n${GREEN}Starting rate limiter test with %d requests per IP${NC}\n" "$REQUESTS_PER_IP"
printf "Testing IP addresses: %s\n" "${IPS[*]}"
printf "Execution time range: 50-100ms\n\n"

# Основной цикл тестирования
for ((i=1; i<=$REQUESTS_PER_IP; i++)); do
    printf "\nBatch %d/%d\n" "$i" "$REQUESTS_PER_IP"
    for ip in "${IPS[@]}"; do
        send_request "$ip"
        sleep $DELAY
    done
done

# Вывод статистики
printf "\n${GREEN}Test completed. Statistics:${NC}\n"
for ip in "${IPS[@]}"; do
    success=$(get_counter "$ip" "success")
    limit=$(get_counter "$ip" "limit")
    error=$(get_counter "$ip" "error")
    total=$((success + limit + error))

    if [ $total -gt 0 ]; then
        success_pct=$((success * 100 / total))
        limit_pct=$((limit * 100 / total))
        error_pct=$((error * 100 / total))
    else
        success_pct=0
        limit_pct=0
        error_pct=0
    fi

    printf "\nIP: %s\n" "$ip"
    printf "Total requests: %d\n" "$total"
    printf "${GREEN}Success: %d (%d%%)${NC}\n" "$success" "$success_pct"
    printf "${YELLOW}Rate Limited: %d (%d%%)${NC}\n" "$limit" "$limit_pct"
    printf "${RED}Errors: %d (%d%%)${NC}\n" "$error" "$error_pct"
done
