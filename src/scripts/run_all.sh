#!/bin/bash

CONFIG_FILE="config/servers.json"
LOG_DIR="logs"
PID_FILE="$LOG_DIR/server.pids"

mkdir -p "$LOG_DIR"

# Очистка PID файла
> "$PID_FILE"

# Функция для добавления PID в файл
add_pid() {
    echo "$1" >> "$PID_FILE"
}

# Функция для инициализации лог-файла
init_log() {
    local log_file=$1
    local service_name=$2
    local port=$3
    
    if [ ! -f "$log_file" ]; then
        touch "$log_file"
    fi
    echo -e "\n\n=== [$(date '+%Y-%m-%d %H:%M:%S')] New session for $service_name on port $port ===" >> "$log_file"
}

# Функция для проверки занятости порта
is_port_free() {
    local port=$1
    if lsof -i :"$port" >/dev/null 2>&1; then
        echo "Port $port is already in use. Trying to free..."
        local pids=$(lsof -ti :"$port")
        if [ -n "$pids" ]; then
            kill -9 $pids 2>/dev/null
            sleep 1
            if lsof -i :"$port" >/dev/null 2>&1; then
                echo "Failed to free port $port"
                return 1
            fi
        fi
    fi
    return 0
}

# Запуск серверов
jq -c '.[]' "$CONFIG_FILE" | while read -r server; do
    id=$(echo "$server" | jq -r '.id')
    port=$(echo "$server" | jq -r '.port')
    log_file="$LOG_DIR/backend_$port.log"
    
    init_log "$log_file" "backend" "$port"
    
    if ! is_port_free "$port"; then
        echo "ERROR: Could not free port $port for server $id" >> "$log_file"
        continue
    fi

    echo "Starting server $id on port $port (logs: $log_file)"
    go run cmd/backend/main.go "$CONFIG_FILE" "$id" >> "$log_file" 2>&1 &
    add_pid $!
    
    sleep 1
    
    if ! ps -p $! > /dev/null; then
        echo "ERROR: Server $id failed to start" >> "$log_file"
    fi
done

# Запуск балансировщика
balancer_port="8080"
balancer_log="$LOG_DIR/balancer_$balancer_port.log"
init_log "$balancer_log" "balancer" "$balancer_port"

echo "Starting load balancer (logs: $balancer_log)"
go run cmd/balancer/main.go >> "$balancer_log" 2>&1 &
add_pid $!

# Функция для корректного завершения
cleanup() {
    echo -e "\n[$(date '+%Y-%m-%d %H:%M:%S')] Stopping all processes..." | tee -a "$balancer_log"
    if [ -f "$PID_FILE" ]; then
        while read pid; do
            kill "$pid" 2>/dev/null
        done < "$PID_FILE"
    fi
    rm -f "$PID_FILE"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

wait