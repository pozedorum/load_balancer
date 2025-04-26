#!/bin/bash

for i in {1..5}; do
  exec_time=$(( RANDOM % 4000 + 1500 ))
  curl -s -w "%{http_code}" -H "Execution-Time: $exec_time" http://localhost:8080
  echo "Execution-Time: $exec_time"
done