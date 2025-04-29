#!/bin/bash

# Настройки теста
BALANCER_URL="http://localhost:8080"
REQUESTS=10
MAX_DELAY=3000

echo "Starting load balancer test suite..."
echo "===================================="

# Функция для анализа ответов
process_response() {
    local response="$1"
    if [[ "$response" == *"server_id"* ]]; then
        # Парсим JSON без jq
        server_id=$(echo "$response" | grep -o '"server_id":[0-9]*' | cut -d: -f2)
        delay=$(echo "$response" | grep -o '"delay":[0-9]*' | cut -d: -f2)
        echo "$server_id,$delay"
    else
        echo "Error: $response"
    fi
}

# 1. Тест с равномерной нагрузкой
echo -e "\nTest 1: Evenly distributed requests"
results1=""
for ((i=1; i<=$REQUESTS; i++)); do
    delay=$((i * 200))
    echo "Request $i with delay ${delay}ms"
    response=$(curl -s -H "Execution-Time: $delay" "$BALANCER_URL/process?test=1")
    results1+=$(process_response "$response")$'\n'
done

# Статистика для Test 1
echo "$results1" | grep -v "Error" | awk -F, '{
    servers[$1]++; 
    sum[$1]+=$2
} 
END {
    for (s in servers) 
        printf "Server %s: %d requests, avg delay %.2fms\n", s, servers[s], sum[s]/servers[s]/1000000
}'

# 2. Тест со случайными задержками
echo -e "\nTest 2: Random delays"
results2=""
for ((i=1; i<=$REQUESTS; i++)); do
    delay=$((RANDOM % MAX_DELAY))
    echo "Request $i with random delay ${delay}ms"
    response=$(curl -s -H "Execution-Time: $delay" "$BALANCER_URL/process?test=2")
    results2+=$(process_response "$response")$'\n'
done

# Статистика для Test 2
echo "$results2" | grep -v "Error" | awk -F, '{
    servers[$1]++; 
    sum[$1]+=$2
} 
END {
    for (s in servers) 
        printf "Server %s: %d requests, avg delay %.2fms\n", s, servers[s], sum[s]/servers[s]/1000000
}'

# 3. Тест с ошибками
echo -e "\nTest 3: Error cases"
curl -s -H "Execution-Time: invalid" "$BALANCER_URL/process?test=3"
curl -s -H "Execution-Time: -100" "$BALANCER_URL/process?test=3"
curl -s -H "Execution-Time: 100000" "$BALANCER_URL/process?test=3"

echo -e "\nTest completed"