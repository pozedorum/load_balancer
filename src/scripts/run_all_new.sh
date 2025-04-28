#!/bin/bash

CONFIG_FILE="config/servers.json"

# Запускаем бэкенды
jq -c '.[]' $CONFIG_FILE | while read server; do
    id=$(echo "$server" | jq -r '.id')
    port=$(echo "$server" | jq -r '.port')
    
    echo "Starting server $id on port $port"
    if ! go run cmd/backend/main.go $CONFIG_FILE $id & then
        echo "Failed to start server $id"
        exit 1
    fi
    sleep 1
done

# Запускаем балансировщик
echo "Starting load balancer"
go run cmd/balancer/main.go

# Остановка всех процессов при завершении
trap "pkill -P $$" EXIT