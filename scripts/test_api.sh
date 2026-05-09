#!/bin/bash

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

# Helper function to print test results
print_result() {
    local test_name=$1
    local status_code=$2
    local expected_code=$3

    if [ "$status_code" -eq "$expected_code" ]; then
        echo -e "${test_name}: ${GREEN}PASS${NC} (Status: $status_code)"
    else
        echo -e "${test_name}: ${RED}FAIL${NC} (Expected: $expected_code, Got: $status_code)"
    fi
}

echo "Running ProxyMaze'26 API Smoke Tests..."
echo "========================================="

# 1. GET /health
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/health)
print_result "1. GET /health" $status 200

# 2. POST /config
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/config \
  -H "Content-Type: application/json" \
  -d '{"check_interval_seconds": 1, "request_timeout_ms": 3000}')
print_result "2. POST /config" $status 200

# 3. GET /config
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/config)
print_result "3. GET /config" $status 200

# 4. POST /proxies
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/proxies \
  -H "Content-Type: application/json" \
  -d '{"proxies": ["http://google.com/proxy1", "http://github.com/proxy2"], "replace": true}')
status=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result "4. POST /proxies" $status 201

# Extract the first proxy ID to use in subsequent requests
PROXY_ID=$(echo $body | grep -o '"id":"[^"]*' | head -1 | awk -F'"' '{print $4}')
if [ -z "$PROXY_ID" ]; then
  PROXY_ID="proxy1" # fallback
fi

# 5. GET /proxies
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/proxies)
print_result "5. GET /proxies" $status 200

# Wait briefly for a background check to happen
sleep 2

# 6. GET /proxies/{id}
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/proxies/$PROXY_ID)
print_result "6. GET /proxies/{id}" $status 200

# 7. GET /proxies/{id}/history
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/proxies/$PROXY_ID/history)
print_result "7. GET /proxies/{id}/history" $status 200

# 8. GET /alerts
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/alerts)
print_result "8. GET /alerts" $status 200

# 9. POST /webhooks
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/webhooks \
  -H "Content-Type: application/json" \
  -d '{"url": "http://localhost:8080/dummy-webhook"}')
print_result "9. POST /webhooks" $status 201

# 10. POST /integrations (slack)
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/integrations \
  -H "Content-Type: application/json" \
  -d '{"type": "slack", "webhook_url": "http://dummy", "username": "ProxyWatch", "events": ["alert.fired","alert.resolved"]}')
print_result "10. POST /integrations (slack)" $status 201

# 11. POST /integrations (discord)
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/integrations \
  -H "Content-Type: application/json" \
  -d '{"type": "discord", "webhook_url": "http://dummy", "username": "ProxyWatch", "events": ["alert.fired","alert.resolved"]}')
print_result "11. POST /integrations (discord)" $status 201

# 12. DELETE /proxies
status=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE_URL/proxies)
print_result "12. DELETE /proxies" $status 204

# 13. GET /metrics
status=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/metrics)
print_result "13. GET /metrics" $status 200

echo "========================================="
echo "Done!"
