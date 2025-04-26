#!/bin/bash

# Останавливаем все процессы из pids.txt
if [ -f pids.txt ]; then
    while read -r pid; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            echo "Остановлен процесс $pid"
        fi
    done < pids.txt
    rm pids.txt
fi

echo "Все сервисы остановлены"