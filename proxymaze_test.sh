#!/usr/bin/env bash
# =============================================================================
# ProxyMaze'26 — Black-Box Test Suite
# Run against your live service: BASE_URL=http://localhost:8080 bash proxymaze_test.sh
# Requires: curl, jq, nc (netcat), python3 (for the mock webhook server)
# =============================================================================

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0
WARN=0

# ── Color helpers ─────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

pass()  { echo -e "${GREEN}  ✓ PASS${NC}  $1"; ((PASS++)); }
fail()  { echo -e "${RED}  ✗ FAIL${NC}  $1"; ((FAIL++)); }
warn()  { echo -e "${YELLOW}  ⚠ WARN${NC}  $1"; ((WARN++)); }
section(){ echo -e "\n${CYAN}${BOLD}══ $1 ══${NC}"; }

# ── Assertion helpers ─────────────────────────────────────────────────────────
assert_status() {
  local label="$1" expected="$2" actual="$3"
  [[ "$actual" == "$expected" ]] && pass "$label (HTTP $actual)" || fail "$label — expected HTTP $expected, got $actual"
}

assert_json_field() {
  local label="$1" body="$2" path="$3" expected="$4"
  local actual; actual=$(echo "$body" | jq -r "$path" 2>/dev/null)
  [[ "$actual" == "$expected" ]] && pass "$label ($path = $actual)" \
    || fail "$label — expected $path=$expected, got $actual"
}

assert_json_not_null() {
  local label="$1" body="$2" path="$3"
  local actual; actual=$(echo "$body" | jq -r "$path" 2>/dev/null)
  [[ -n "$actual" && "$actual" != "null" ]] && pass "$label ($path present)" \
    || fail "$label — $path is null or missing"
}

assert_json_is_array() {
  local label="$1" body="$2"
  echo "$body" | jq -e 'type == "array"' >/dev/null 2>&1 \
    && pass "$label (is JSON array)" || fail "$label — body is not a JSON array"
}

assert_contains() {
  local label="$1" haystack="$2" needle="$3"
  [[ "$haystack" == *"$needle"* ]] && pass "$label (contains '$needle')" \
    || fail "$label — '$needle' not found in: $haystack"
}

# ── HTTP helpers ──────────────────────────────────────────────────────────────
GET()  { curl -s -w "\n%{http_code}" "$BASE_URL$1"; }
POST() { curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$2" "$BASE_URL$1"; }
DEL()  { curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL$1"; }

split_body_code() {
  # Last line is status code, everything else is body
  local raw="$1"
  CODE=$(echo "$raw" | tail -1)
  BODY=$(echo "$raw" | head -n -1)
}

sleep_msg() { echo -e "${YELLOW}  ⏳ Waiting $1s — $2${NC}"; sleep "$1"; }

# =============================================================================
# MOCK WEBHOOK SERVER (Python background process)
# Listens on port 19876, records payloads to /tmp/pmaze_webhooks.log
# =============================================================================
WEBHOOK_PORT=19876
WEBHOOK_LOG=/tmp/pmaze_webhooks.log
WEBHOOK_PID=""

start_webhook_server() {
  rm -f "$WEBHOOK_LOG"
  python3 - "$WEBHOOK_PORT" "$WEBHOOK_LOG" &
  WEBHOOK_PID=$!
  sleep 0.5
  cat <<'PYEOF' >/dev/null
# (inline, already launched above via heredoc trick — see below)
PYEOF
}

# Launch mock webhook via a Python one-liner
python3 -c "
import sys, http.server, json, time

port = int(sys.argv[1])
logfile = sys.argv[2]

class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(length).decode()
        with open(logfile, 'a') as f:
            f.write(json.dumps({'ts': time.time(), 'path': self.path, 'body': body}) + '\n')
        self.send_response(200)
        self.end_headers()
    def log_message(self, *a): pass

http.server.HTTPServer(('0.0.0.0', port), H).serve_forever()
" "$WEBHOOK_PORT" "$WEBHOOK_LOG" &
WEBHOOK_PID=$!
sleep 0.6
echo -e "${CYAN}Mock webhook server started on :$WEBHOOK_PORT (PID $WEBHOOK_PID)${NC}"
WEBHOOK_URL="http://host.docker.internal:${WEBHOOK_PORT}/hook"
# Fallback for non-Docker
if ! curl -s --connect-timeout 1 "http://host.docker.internal:${WEBHOOK_PORT}" >/dev/null 2>&1; then
  WEBHOOK_URL="http://localhost:${WEBHOOK_PORT}/hook"
fi

cleanup() {
  [[ -n "$WEBHOOK_PID" ]] && kill "$WEBHOOK_PID" 2>/dev/null
}
trap cleanup EXIT

# =============================================================================
# HELPER: wait for a proxy status to appear
# =============================================================================
wait_for_proxy_status() {
  local proxy_id="$1" expected_status="$2" max_wait="${3:-30}"
  local elapsed=0
  while (( elapsed < max_wait )); do
    local raw; raw=$(GET "/proxies/$proxy_id")
    split_body_code "$raw"
    local status; status=$(echo "$BODY" | jq -r '.status' 2>/dev/null)
    [[ "$status" == "$expected_status" ]] && return 0
    sleep 2; (( elapsed += 2 ))
  done
  return 1
}

# Wait until pool failure_rate crosses a threshold
wait_for_failure_rate() {
  local op="$1" threshold="$2" max_wait="${3:-40}"
  local elapsed=0
  while (( elapsed < max_wait )); do
    local raw; raw=$(GET "/proxies")
    split_body_code "$raw"
    local rate; rate=$(echo "$BODY" | jq -r '.failure_rate' 2>/dev/null)
    if python3 -c "import sys; r=float('${rate:-0}'); sys.exit(0 if r ${op} ${threshold} else 1)" 2>/dev/null; then
      return 0
    fi
    sleep 2; (( elapsed += 2 ))
  done
  return 1
}

# Count webhook deliveries matching an event type
count_webhook_events() {
  local event="$1"
  grep -c "\"event\": *\"$event\"" "$WEBHOOK_LOG" 2>/dev/null || echo 0
}

# =============================================================================
# TEST PROXIES (use httpbin-style URLs — always-up and always-down)
# =============================================================================
GOOD_URL="https://httpbin.org/status/200"    # always returns 200
BAD_URL="https://httpbin.org/status/500"     # always returns 500
TIMEOUT_URL="https://10.255.255.1/timeout"   # unreachable — causes timeout

# For the pool-breach tests we need >20% down.
# We'll use a 5-proxy pool: 2 down = 40% failure rate.

# =============================================================================
# PART 0 — RESET STATE
# =============================================================================
section "RESET — Clear pool before tests"
raw=$(DEL "/proxies"); split_body_code "$raw"
# Accept 204 or 200
[[ "$CODE" == "204" || "$CODE" == "200" ]] && pass "DELETE /proxies reset" || warn "DELETE /proxies returned $CODE (expected 204)"

# Reset config to fast interval for tests
raw=$(POST "/config" '{"check_interval_seconds":5,"request_timeout_ms":4000}')
split_body_code "$raw"
assert_status "POST /config (fast interval for tests)" "200" "$CODE"


# =============================================================================
# CHAPTER 01 — GET /health
# =============================================================================
section "CH01 · Proof of Life — GET /health"

raw=$(GET "/health"); split_body_code "$raw"
assert_status "GET /health returns 200" "200" "$CODE"
assert_json_field "status field = ok" "$BODY" ".status" "ok"


# =============================================================================
# CHAPTER 02 & 03 — POST /config + GET /config
# =============================================================================
section "CH02/03 · Heartbeat & Memory — POST+GET /config"

raw=$(POST "/config" '{"check_interval_seconds":10,"request_timeout_ms":2000}')
split_body_code "$raw"
assert_status "POST /config returns 200" "200" "$CODE"

raw=$(GET "/config"); split_body_code "$raw"
assert_status "GET /config returns 200" "200" "$CODE"
assert_json_field "check_interval_seconds remembered" "$BODY" ".check_interval_seconds" "10"
assert_json_field "request_timeout_ms remembered" "$BODY" ".request_timeout_ms" "2000"

# Unknown fields must be ignored
raw=$(POST "/config" '{"check_interval_seconds":5,"request_timeout_ms":4000,"unknown_field":"should_be_ignored"}')
split_body_code "$raw"
assert_status "POST /config with unknown field still returns 200" "200" "$CODE"

raw=$(GET "/config"); split_body_code "$raw"
assert_json_field "check_interval_seconds after unknown-field POST" "$BODY" ".check_interval_seconds" "5"


# =============================================================================
# CHAPTER 04 — POST /proxies
# =============================================================================
section "CH04 · Building the Pool — POST /proxies"

# Basic load
raw=$(POST "/proxies" "{\"proxies\":[\"$GOOD_URL\",\"${GOOD_URL}2\"],\"replace\":true}")
split_body_code "$raw"
assert_status "POST /proxies returns 201" "201" "$CODE"
assert_json_field "accepted count = 2" "$BODY" ".accepted" "2"

first_proxy_id=$(echo "$BODY" | jq -r '.proxies[0].id')
assert_json_field "proxies[0].status = pending" "$BODY" ".proxies[0].status" "pending"
assert_json_not_null "proxies[0].id present" "$BODY" ".proxies[0].id"
assert_json_not_null "proxies[0].url present" "$BODY" ".proxies[0].url"

# Proxy ID extraction rule: last path segment
raw=$(POST "/proxies" '{"proxies":["https://proxy-provider.example/proxy/px-101"],"replace":true}')
split_body_code "$raw"
assert_json_field "proxy ID = last path segment (px-101)" "$BODY" ".proxies[0].id" "px-101"

# replace=false should append
raw=$(POST "/proxies" '{"proxies":["https://proxy-provider.example/proxy/px-102"],"replace":false}')
split_body_code "$raw"
assert_status "POST /proxies replace=false returns 201" "201" "$CODE"

raw=$(GET "/proxies"); split_body_code "$raw"
total=$(echo "$BODY" | jq -r '.total')
[[ "$total" -ge 2 ]] && pass "replace=false appended (total=$total)" || fail "replace=false did not append (total=$total)"

# replace=true clears pool
raw=$(POST "/proxies" '{"proxies":["https://proxy-provider.example/proxy/px-201"],"replace":true}')
split_body_code "$raw"
raw=$(GET "/proxies"); split_body_code "$raw"
assert_json_field "replace=true cleared old proxies (total=1)" "$BODY" ".total" "1"

# Unknown fields in POST /proxies must not fail
raw=$(POST "/proxies" '{"proxies":["https://proxy-provider.example/proxy/px-300"],"replace":true,"metadata":"ignored"}')
split_body_code "$raw"
assert_status "POST /proxies with unknown fields still 201" "201" "$CODE"


# =============================================================================
# CHAPTER 05 — GET /proxies
# =============================================================================
section "CH05 · Watchtower — GET /proxies"

# Load a clean pool for this test
raw=$(POST "/proxies" "{\"proxies\":[\"$GOOD_URL\",\"$BAD_URL\"],\"replace\":true}")
split_body_code "$raw"

sleep_msg 12 "Waiting for background checks to run"

raw=$(GET "/proxies"); split_body_code "$raw"
assert_status "GET /proxies returns 200" "200" "$CODE"
assert_json_not_null "total field present" "$BODY" ".total"
assert_json_not_null "up field present" "$BODY" ".up"
assert_json_not_null "down field present" "$BODY" ".down"
assert_json_not_null "failure_rate field present" "$BODY" ".failure_rate"
assert_json_not_null "proxies array present" "$BODY" ".proxies"

# Each proxy entry must have required fields
first=$(echo "$BODY" | jq '.proxies[0]')
for field in id url status last_checked_at consecutive_failures; do
  val=$(echo "$first" | jq -r ".$field")
  [[ -n "$val" && "$val" != "null" ]] \
    && pass "proxies[0].$field present" \
    || fail "proxies[0].$field missing (got: $val)"
done

# Values must not be stale (last_checked_at should be within the last 2 minutes)
last_checked=$(echo "$first" | jq -r '.last_checked_at')
if [[ -n "$last_checked" && "$last_checked" != "null" ]]; then
  ts_now=$(date -u +%s)
  ts_checked=$(python3 -c "from datetime import datetime; print(int(datetime.fromisoformat('${last_checked%Z}').timestamp()))" 2>/dev/null || echo 0)
  diff=$(( ts_now - ts_checked ))
  [[ $diff -lt 120 ]] && pass "last_checked_at is recent ($diff s ago)" \
    || fail "last_checked_at is stale ($diff s ago) — background monitor may not be running"
fi

# failure_rate arithmetic: down/total
total=$(echo "$BODY" | jq -r '.total')
down_count=$(echo "$BODY" | jq -r '.down')
failure_rate=$(echo "$BODY" | jq -r '.failure_rate')
expected_rate=$(python3 -c "print(round($down_count/$total,10))" 2>/dev/null)
python3 -c "abs(float('$failure_rate') - float('$expected_rate')) < 0.001 or exit(1)" 2>/dev/null \
  && pass "failure_rate = down/total ($down_count/$total ≈ $failure_rate)" \
  || fail "failure_rate mismatch: got $failure_rate, expected $expected_rate"


# =============================================================================
# CHAPTER 06 — GET /proxies/{id}
# =============================================================================
section "CH06 · Dossier — GET /proxies/{id}"

proxy_id=$(echo "$BODY" | jq -r '.proxies[0].id')
raw=$(GET "/proxies/$proxy_id"); split_body_code "$raw"
assert_status "GET /proxies/{id} returns 200" "200" "$CODE"

for field in id url status last_checked_at consecutive_failures total_checks uptime_percentage history; do
  val=$(echo "$BODY" | jq -r ".$field")
  [[ -n "$val" && "$val" != "null" ]] \
    && pass "GET /proxies/{id} — $field present" \
    || fail "GET /proxies/{id} — $field missing (got: $val)"
done

# history must be an array
assert_json_is_array "GET /proxies/{id} .history is array" "$(echo "$BODY" | jq '.history')"

# uptime_percentage sanity check (0–100)
pct=$(echo "$BODY" | jq -r '.uptime_percentage')
python3 -c "0 <= float('${pct:-0}') <= 100 or exit(1)" 2>/dev/null \
  && pass "uptime_percentage in range [0,100] ($pct)" \
  || fail "uptime_percentage out of range ($pct)"

# 404 for unknown ID
raw=$(GET "/proxies/does-not-exist-xyz"); split_body_code "$raw"
assert_status "GET /proxies/{unknown} returns 404" "404" "$CODE"


# =============================================================================
# CHAPTER 07 — GET /proxies/{id}/history
# =============================================================================
section "CH07 · Chronicle — GET /proxies/{id}/history"

raw=$(GET "/proxies/$proxy_id/history"); split_body_code "$raw"
assert_status "GET /proxies/{id}/history returns 200" "200" "$CODE"
assert_json_is_array "history body is JSON array" "$BODY"

entry=$(echo "$BODY" | jq '.[0]' 2>/dev/null)
if [[ -n "$entry" && "$entry" != "null" ]]; then
  assert_json_not_null "history[0].checked_at present" "$entry" ".checked_at"
  assert_json_not_null "history[0].status present"     "$entry" ".status"
fi

# 404 for unknown ID
raw=$(GET "/proxies/nonexistent-abc/history"); split_body_code "$raw"
assert_status "GET /proxies/{unknown}/history returns 404" "404" "$CODE"


# =============================================================================
# CHAPTER 08 — DELETE /proxies
# =============================================================================
section "CH08 · Graveyard — DELETE /proxies"

# First fire an alert so we have alert history to preserve
raw=$(POST "/proxies" "{\"proxies\":[\"$BAD_URL\",\"$BAD_URL/2\",\"$BAD_URL/3\",\"$BAD_URL/4\",\"$GOOD_URL\"],\"replace\":true}")
split_body_code "$raw"
sleep_msg 15 "Waiting for breach to fire before DELETE /proxies test"

raw=$(DEL "/proxies"); split_body_code "$raw"
assert_status "DELETE /proxies returns 204" "204" "$CODE"

raw=$(GET "/proxies"); split_body_code "$raw"
total=$(echo "$BODY" | jq -r '.total')
assert_json_field "pool is empty after DELETE (total=0)" "$BODY" ".total" "0"

# Alert history must survive
raw=$(GET "/alerts"); split_body_code "$raw"
assert_status "GET /alerts after DELETE /proxies still 200" "200" "$CODE"
assert_json_is_array "alerts still present after pool clear" "$BODY"


# =============================================================================
# CHAPTER 09 — GET /alerts  (Alert Archive)
# =============================================================================
section "CH09 · Alert Archive — GET /alerts"

raw=$(GET "/alerts"); split_body_code "$raw"
assert_status "GET /alerts returns 200" "200" "$CODE"
assert_json_is_array "GET /alerts body is JSON array" "$BODY"

# If there are any alerts, validate required fields
count=$(echo "$BODY" | jq 'length')
if [[ "$count" -gt 0 ]]; then
  alert=$(echo "$BODY" | jq '.[0]')
  for field in alert_id status failure_rate total_proxies failed_proxies failed_proxy_ids threshold fired_at message; do
    val=$(echo "$alert" | jq -r ".$field")
    [[ -n "$val" && "$val" != "null" ]] \
      && pass "alerts[0].$field present" \
      || fail "alerts[0].$field missing or null"
  done
  assert_json_field "threshold = 0.2" "$alert" ".threshold" "0.2"
  # failed_proxy_ids must be an array
  assert_json_is_array "failed_proxy_ids is array" "$(echo "$alert" | jq '.failed_proxy_ids')"
fi


# =============================================================================
# CHAPTER 10 — POST /webhooks
# =============================================================================
section "CH10 · Messenger — POST /webhooks"

raw=$(POST "/webhooks" "{\"url\":\"$WEBHOOK_URL\"}")
split_body_code "$raw"
assert_status "POST /webhooks returns 201" "201" "$CODE"
assert_json_not_null "webhook_id present" "$BODY" ".webhook_id"
assert_json_field "webhook url echoed" "$BODY" ".url" "$WEBHOOK_URL"

WEBHOOK_ID=$(echo "$BODY" | jq -r '.webhook_id')

# Unknown fields must be accepted
raw=$(POST "/webhooks" "{\"url\":\"${WEBHOOK_URL}2\",\"description\":\"ignored\"}")
split_body_code "$raw"
assert_status "POST /webhooks with unknown fields still 201" "201" "$CODE"


# =============================================================================
# FULL ALERT LIFECYCLE TEST
# =============================================================================
section "ALERT LIFECYCLE — Breach → Fire → Resolve → Re-breach"

# Clear everything
DEL "/proxies" >/dev/null
sleep 1

# Load 5 proxies: 3 bad (60% failure), 2 good — well above 20% threshold
raw=$(POST "/proxies" "{
  \"proxies\": [
    \"$BAD_URL\",
    \"${BAD_URL}/p2\",
    \"${BAD_URL}/p3\",
    \"$GOOD_URL\",
    \"${GOOD_URL}/p5\"
  ],
  \"replace\": true
}")
split_body_code "$raw"
assert_status "POST /proxies (5-proxy breach pool) returns 201" "201" "$CODE"

echo ""
echo -e "${YELLOW}Waiting up to 40s for alert to fire...${NC}"
if wait_for_failure_rate ">=" 0.20 40; then
  pass "Failure rate reached ≥ 0.20"
else
  warn "Failure rate never reached 0.20 in 40s — proxy probe may be slow or bad URLs aren't failing"
fi

sleep_msg 10 "Giving webhook dispatcher time to deliver alert.fired"

# ── Validate alert object ────────────────────────────────────────────────────
raw=$(GET "/alerts"); split_body_code "$raw"
active_alerts=$(echo "$BODY" | jq '[.[] | select(.status=="active")]')
active_count=$(echo "$active_alerts" | jq 'length')

[[ "$active_count" -eq 1 ]] && pass "Exactly 1 active alert" \
  || fail "Expected 1 active alert, found $active_count"

if [[ "$active_count" -ge 1 ]]; then
  alert=$(echo "$active_alerts" | jq '.[0]')
  ALERT_ID=$(echo "$alert" | jq -r '.alert_id')
  assert_json_not_null  "alert_id present"            "$alert" ".alert_id"
  assert_json_field     "alert status = active"       "$alert" ".status"        "active"
  assert_json_field     "threshold = 0.2"             "$alert" ".threshold"     "0.2"
  assert_json_not_null  "fired_at present"            "$alert" ".fired_at"
  assert_json_field     "resolved_at = null"          "$alert" ".resolved_at"   "null"
  assert_json_not_null  "failed_proxy_ids non-empty"  "$alert" ".failed_proxy_ids[0]"
  assert_json_not_null  "message non-empty"           "$alert" ".message"

  # Cross-check: GET /proxies and GET /alerts must agree on failed set
  raw2=$(GET "/proxies"); split_body_code "$raw2"
  proxies_down_ids=$(echo "$BODY" | jq -r '[.proxies[] | select(.status=="down") | .id] | sort | join(",")' 2>/dev/null)
  alert_failed_ids=$(echo "$alert" | jq -r '.failed_proxy_ids | sort | join(",")' 2>/dev/null)
  [[ "$proxies_down_ids" == "$alert_failed_ids" ]] \
    && pass "GET /proxies and GET /alerts agree on failed_proxy_ids ($alert_failed_ids)" \
    || fail "Mismatch: /proxies down=$proxies_down_ids vs /alerts failed_ids=$alert_failed_ids"

  # total_proxies consistency
  proxies_total=$(echo "$BODY" | jq -r '.total')
  alert_total=$(echo "$alert" | jq -r '.total_proxies')
  [[ "$proxies_total" == "$alert_total" ]] \
    && pass "total_proxies consistent (/proxies=$proxies_total, alert=$alert_total)" \
    || fail "total_proxies mismatch: /proxies=$proxies_total vs alert=$alert_total"
fi

# ── Validate webhook delivery of alert.fired ─────────────────────────────────
fired_count=$(count_webhook_events "alert.fired")
[[ "$fired_count" -ge 1 ]] && pass "alert.fired webhook delivered ($fired_count event(s))" \
  || fail "alert.fired webhook NOT delivered (check mock webhook log: $WEBHOOK_LOG)"

[[ "$fired_count" -eq 1 ]] && pass "alert.fired delivered exactly once (no duplicates)" \
  || warn "alert.fired delivered $fired_count times — possible duplicate deliveries"

# Validate alert.fired payload fields
if [[ "$fired_count" -ge 1 ]]; then
  fired_payload=$(grep "alert.fired" "$WEBHOOK_LOG" | tail -1 | python3 -c "import sys,json; line=json.loads(sys.stdin.read()); print(line['body'])" 2>/dev/null)
  for field in event alert_id fired_at failure_rate total_proxies failed_proxies failed_proxy_ids threshold message; do
    val=$(echo "$fired_payload" | jq -r ".$field" 2>/dev/null)
    [[ -n "$val" && "$val" != "null" ]] \
      && pass "webhook alert.fired payload: $field present" \
      || fail "webhook alert.fired payload: $field missing"
  done
fi

# ── No duplicate active alerts (breach continues) ───────────────────────────
sleep_msg 8 "Checking no duplicate alert fires during sustained breach"
raw=$(GET "/alerts"); split_body_code "$raw"
active_count2=$(echo "$BODY" | jq '[.[] | select(.status=="active")] | length')
[[ "$active_count2" -eq 1 ]] && pass "Still exactly 1 active alert during sustained breach" \
  || fail "Duplicate alerts fired during sustained breach (count=$active_count2)"

fired_count2=$(count_webhook_events "alert.fired")
[[ "$fired_count2" -eq 1 ]] && pass "No duplicate alert.fired webhooks during sustained breach" \
  || fail "Duplicate alert.fired webhooks: $fired_count2 deliveries during sustained breach"

# =============================================================================
# RESOLVE — Replace pool with all-good proxies
# =============================================================================
section "ALERT RESOLVE — Pool recovery"

raw=$(POST "/proxies" "{
  \"proxies\": [
    \"$GOOD_URL\",
    \"${GOOD_URL}/p2\",
    \"${GOOD_URL}/p3\",
    \"${GOOD_URL}/p4\",
    \"${GOOD_URL}/p5\"
  ],
  \"replace\": true
}")
split_body_code "$raw"
assert_status "POST /proxies (all-good recovery pool) 201" "201" "$CODE"

echo -e "${YELLOW}Waiting up to 40s for alert to resolve...${NC}"
resolved=false
for i in $(seq 1 20); do
  raw=$(GET "/alerts"); split_body_code "$raw"
  active=$(echo "$BODY" | jq '[.[] | select(.status=="active")] | length')
  if [[ "$active" -eq 0 ]]; then
    resolved=true; break
  fi
  sleep 2
done
$resolved && pass "Alert resolved when failure rate dropped below 0.20" \
  || fail "Alert NOT resolved after 40s with all-good pool"

sleep_msg 8 "Waiting for alert.resolved webhook delivery"

# Validate resolved alert fields
raw=$(GET "/alerts"); split_body_code "$raw"
resolved_alerts=$(echo "$BODY" | jq '[.[] | select(.status=="resolved")]')
resolved_count=$(echo "$resolved_alerts" | jq 'length')
[[ "$resolved_count" -ge 1 ]] && pass "At least 1 resolved alert in archive" \
  || fail "No resolved alerts found"

if [[ "$resolved_count" -ge 1 ]]; then
  resolved_alert=$(echo "$resolved_alerts" | jq '.[0]')
  assert_json_not_null "resolved_at populated on resolved alert" "$resolved_alert" ".resolved_at"
  assert_json_field    "resolved alert status = resolved"         "$resolved_alert" ".status" "resolved"
fi

# Validate alert.resolved webhook
resolved_webhook_count=$(count_webhook_events "alert.resolved")
[[ "$resolved_webhook_count" -ge 1 ]] && pass "alert.resolved webhook delivered" \
  || fail "alert.resolved webhook NOT delivered"

# Validate alert.resolved payload
if [[ "$resolved_webhook_count" -ge 1 ]]; then
  resolved_payload=$(grep "alert.resolved" "$WEBHOOK_LOG" | tail -1 | python3 -c "import sys,json; line=json.loads(sys.stdin.read()); print(line['body'])" 2>/dev/null)
  for field in event alert_id resolved_at; do
    val=$(echo "$resolved_payload" | jq -r ".$field" 2>/dev/null)
    [[ -n "$val" && "$val" != "null" ]] \
      && pass "webhook alert.resolved payload: $field present" \
      || fail "webhook alert.resolved payload: $field missing"
  done
  # alert_id in resolved must match the fired alert
  if [[ -n "$ALERT_ID" ]]; then
    resolved_aid=$(echo "$resolved_payload" | jq -r '.alert_id' 2>/dev/null)
    [[ "$resolved_aid" == "$ALERT_ID" ]] \
      && pass "alert.resolved.alert_id matches alert.fired.alert_id ($ALERT_ID)" \
      || fail "alert_id mismatch: fired=$ALERT_ID resolved=$resolved_aid"
  fi
fi

# =============================================================================
# RE-BREACH — New alert must have a brand-new alert_id
# =============================================================================
section "RE-BREACH — New alert_id after recovery"

FIRST_ALERT_ID="$ALERT_ID"

raw=$(POST "/proxies" "{
  \"proxies\": [
    \"$BAD_URL\",
    \"${BAD_URL}/r2\",
    \"${BAD_URL}/r3\",
    \"$GOOD_URL\",
    \"${GOOD_URL}/r5\"
  ],
  \"replace\": true
}")
split_body_code "$raw"

echo -e "${YELLOW}Waiting up to 40s for re-breach alert...${NC}"
new_breach=false
for i in $(seq 1 20); do
  raw=$(GET "/alerts"); split_body_code "$raw"
  new_active=$(echo "$BODY" | jq --arg fid "$FIRST_ALERT_ID" '[.[] | select(.status=="active" and .alert_id != $fid)] | length')
  if [[ "$new_active" -ge 1 ]]; then
    new_breach=true
    NEW_ALERT_ID=$(echo "$BODY" | jq -r --arg fid "$FIRST_ALERT_ID" '[.[] | select(.status=="active" and .alert_id != $fid)][0].alert_id')
    break
  fi
  sleep 2
done
$new_breach && pass "New alert fired on re-breach" \
  || fail "No new alert on re-breach after 40s"

if [[ -n "$NEW_ALERT_ID" && -n "$FIRST_ALERT_ID" ]]; then
  [[ "$NEW_ALERT_ID" != "$FIRST_ALERT_ID" ]] \
    && pass "Re-breach uses new alert_id ($NEW_ALERT_ID ≠ $FIRST_ALERT_ID)" \
    || fail "Re-breach reused old alert_id — must mint a new one"
fi

# Old resolved alert must still be in archive
raw=$(GET "/alerts"); split_body_code "$raw"
old_alert_present=$(echo "$BODY" | jq --arg fid "$FIRST_ALERT_ID" '[.[] | select(.alert_id == $fid)] | length')
[[ "$old_alert_present" -ge 1 ]] && pass "Old resolved alert still in archive after re-breach" \
  || fail "Old resolved alert disappeared from archive after re-breach"

# Total active count still = 1
active_on_rebreach=$(echo "$BODY" | jq '[.[] | select(.status=="active")] | length')
[[ "$active_on_rebreach" -eq 1 ]] && pass "Still only 1 active alert on re-breach" \
  || fail "More than 1 active alert on re-breach (count=$active_on_rebreach)"

sleep_msg 8 "Waiting for re-breach alert.fired webhook"
fired_total=$(count_webhook_events "alert.fired")
[[ "$fired_total" -ge 2 ]] && pass "Second alert.fired webhook delivered for re-breach" \
  || fail "Re-breach alert.fired webhook not delivered (total fired events: $fired_total)"


# =============================================================================
# CHAPTER 11 — POST /integrations  (Slack + Discord)
# =============================================================================
section "CH11 · Integration Layer — POST /integrations"

raw=$(POST "/integrations" "{
  \"type\": \"slack\",
  \"webhook_url\": \"$WEBHOOK_URL\",
  \"username\": \"ProxyWatch\",
  \"events\": [\"alert.fired\", \"alert.resolved\"]
}")
split_body_code "$raw"
[[ "$CODE" == "200" || "$CODE" == "201" ]] && pass "POST /integrations (slack) returns 200 or 201" \
  || fail "POST /integrations (slack) returned $CODE"

raw=$(POST "/integrations" "{
  \"type\": \"discord\",
  \"webhook_url\": \"$WEBHOOK_URL\",
  \"username\": \"ProxyWatch\",
  \"events\": [\"alert.fired\", \"alert.resolved\"]
}")
split_body_code "$raw"
[[ "$CODE" == "200" || "$CODE" == "201" ]] && pass "POST /integrations (discord) returns 200 or 201" \
  || fail "POST /integrations (discord) returned $CODE"

# Unknown fields must not fail
raw=$(POST "/integrations" "{\"type\":\"slack\",\"webhook_url\":\"$WEBHOOK_URL\",\"username\":\"X\",\"events\":[\"alert.fired\"],\"extra\":\"ignored\"}")
split_body_code "$raw"
[[ "$CODE" == "200" || "$CODE" == "201" ]] && pass "POST /integrations with unknown fields accepted" \
  || fail "POST /integrations with unknown fields rejected ($CODE)"


# =============================================================================
# CHAPTER 12 — GET /metrics
# =============================================================================
section "CH12 · Control Room — GET /metrics"

raw=$(GET "/metrics"); split_body_code "$raw"
assert_status "GET /metrics returns 200" "200" "$CODE"

for field in total_checks current_pool_size active_alerts total_alerts webhook_deliveries; do
  val=$(echo "$BODY" | jq -r ".$field" 2>/dev/null)
  [[ -n "$val" && "$val" != "null" ]] \
    && pass "GET /metrics — $field present ($val)" \
    || fail "GET /metrics — $field missing or null"
done

# total_checks should be > 0 if monitoring ran
tc=$(echo "$BODY" | jq -r '.total_checks')
[[ "${tc:-0}" -gt 0 ]] && pass "total_checks > 0 ($tc)" \
  || warn "total_checks = 0 — background monitoring may not be running"


# =============================================================================
# EDGE CASES & BEHAVIORAL RULES
# =============================================================================
section "EDGE CASES"

# Malformed JSON must be rejected
raw=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d '{broken json' "$BASE_URL/config")
split_body_code "$raw"
[[ "$CODE" == "400" ]] && pass "Malformed JSON on POST /config rejected with 400" \
  || warn "Malformed JSON on POST /config returned $CODE (expected 400)"

raw=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d '{broken json' "$BASE_URL/proxies")
split_body_code "$raw"
[[ "$CODE" == "400" ]] && pass "Malformed JSON on POST /proxies rejected with 400" \
  || warn "Malformed JSON on POST /proxies returned $CODE (expected 400)"

# GET /proxies after DELETE must return empty pool (not 404)
DEL "/proxies" >/dev/null
raw=$(GET "/proxies"); split_body_code "$raw"
assert_status "GET /proxies after DELETE still returns 200" "200" "$CODE"
assert_json_field "empty pool total=0" "$BODY" ".total" "0"

# POST /proxies replace omitted defaults to append
raw=$(POST "/proxies" '{"proxies":["https://proxy-provider.example/proxy/append-001"]}')
split_body_code "$raw"
assert_status "POST /proxies without replace field returns 201" "201" "$CODE"

# Timestamps must be ISO 8601 UTC (contain 'T' and 'Z')
raw=$(POST "/proxies" "{\"proxies\":[\"$GOOD_URL\"],\"replace\":true}"); sleep 12
raw=$(GET "/proxies"); split_body_code "$raw"
ts=$(echo "$BODY" | jq -r '.proxies[0].last_checked_at' 2>/dev/null)
if [[ -n "$ts" && "$ts" != "null" ]]; then
  [[ "$ts" == *"T"*"Z" ]] && pass "last_checked_at is ISO 8601 UTC ($ts)" \
    || fail "last_checked_at is NOT ISO 8601 UTC: $ts"
fi

# =============================================================================
# BONUS: SLACK PAYLOAD STRUCTURE CHECK
# =============================================================================
section "BONUS · Slack Integration Payload Validation"

# Re-fire an alert with Slack integration active
raw=$(POST "/proxies" "{
  \"proxies\": [
    \"$BAD_URL\",\"${BAD_URL}/s2\",\"${BAD_URL}/s3\",\"$GOOD_URL\",\"${GOOD_URL}/s5\"
  ],
  \"replace\": true
}")
sleep_msg 20 "Waiting for Slack payload to arrive via webhook"

# Find the latest Slack-style payload (has 'attachments')
slack_payload=$(tac "$WEBHOOK_LOG" 2>/dev/null | python3 -c "
import sys, json
for line in sys.stdin:
    try:
        rec = json.loads(line.strip())
        body = json.loads(rec.get('body','{}'))
        if 'attachments' in body:
            print(rec['body'])
            break
    except: pass
" 2>/dev/null)

if [[ -n "$slack_payload" ]]; then
  pass "Slack-style payload received (has 'attachments')"
  assert_json_not_null "slack.username present"                "$slack_payload" ".username"
  assert_json_not_null "slack.text present"                   "$slack_payload" ".text"
  assert_json_not_null "slack.attachments[0].color present"   "$slack_payload" ".attachments[0].color"
  assert_json_not_null "slack.attachments[0].footer present"  "$slack_payload" ".attachments[0].footer"
  assert_json_not_null "slack.attachments[0].ts present"      "$slack_payload" ".attachments[0].ts"

  # ts must be an integer
  ts_type=$(echo "$slack_payload" | jq '.attachments[0].ts | type' 2>/dev/null)
  [[ "$ts_type" == '"number"' ]] && pass "slack.attachments[0].ts is a number (integer)" \
    || fail "slack.attachments[0].ts is $ts_type — must be integer"

  # color must be #RRGGBB
  color=$(echo "$slack_payload" | jq -r '.attachments[0].color' 2>/dev/null)
  [[ "$color" =~ ^#[0-9A-Fa-f]{6}$ ]] && pass "slack.attachments[0].color is #RRGGBB ($color)" \
    || fail "slack.attachments[0].color format wrong: $color (expected #RRGGBB)"

  # Required field titles
  fields_json=$(echo "$slack_payload" | jq '[.attachments[0].fields[].title | ascii_downcase]' 2>/dev/null)
  for needle in "alert id" "failure rate" "failed proxies" "threshold" "failed ids" "fired at"; do
    python3 -c "
import json, sys
fields = json.loads('''$fields_json''')
needle = '$needle'
found = any(needle in f for f in fields)
sys.exit(0 if found else 1)
" 2>/dev/null && pass "Slack fields include '$needle'" || fail "Slack fields missing '$needle'"
  done
else
  warn "No Slack-style payload found in webhook log — Slack integration may not be delivering"
fi

# =============================================================================
# BONUS: DISCORD PAYLOAD STRUCTURE CHECK
# =============================================================================
section "BONUS · Discord Integration Payload Validation"

discord_payload=$(tac "$WEBHOOK_LOG" 2>/dev/null | python3 -c "
import sys, json
for line in sys.stdin:
    try:
        rec = json.loads(line.strip())
        body = json.loads(rec.get('body','{}'))
        if 'embeds' in body:
            print(rec['body'])
            break
    except: pass
" 2>/dev/null)

if [[ -n "$discord_payload" ]]; then
  pass "Discord-style payload received (has 'embeds')"
  assert_json_not_null "discord.embeds[0].title present"         "$discord_payload" ".embeds[0].title"
  assert_json_not_null "discord.embeds[0].description present"   "$discord_payload" ".embeds[0].description"
  assert_json_not_null "discord.embeds[0].color present"         "$discord_payload" ".embeds[0].color"
  assert_json_not_null "discord.embeds[0].footer.text present"   "$discord_payload" ".embeds[0].footer.text"

  # color must be integer 0–16777215
  d_color=$(echo "$discord_payload" | jq '.embeds[0].color' 2>/dev/null)
  python3 -c "c=int('${d_color:-0}'); assert 0<=c<=16777215" 2>/dev/null \
    && pass "discord.embeds[0].color in range 0–16777215 ($d_color)" \
    || fail "discord.embeds[0].color out of range or non-integer: $d_color"

  # Required field names
  dfields_json=$(echo "$discord_payload" | jq '[.embeds[0].fields[].name | ascii_downcase]' 2>/dev/null)
  for needle in "alert id" "failure rate" "failed proxies" "threshold" "failed ids"; do
    python3 -c "
import json, sys
fields = json.loads('''$dfields_json''')
needle = '$needle'
found = any(needle in f for f in fields)
sys.exit(0 if found else 1)
" 2>/dev/null && pass "Discord fields include '$needle'" || fail "Discord fields missing '$needle'"
  done
else
  warn "No Discord-style payload found in webhook log — Discord integration may not be delivering"
fi


# =============================================================================
# FINAL SCORE SUMMARY
# =============================================================================
section "TEST RESULTS"

TOTAL=$((PASS + FAIL + WARN))
echo ""
echo -e "  Total tests : $TOTAL"
echo -e "  ${GREEN}Passed${NC}      : $PASS"
echo -e "  ${RED}Failed${NC}      : $FAIL"
echo -e "  ${YELLOW}Warnings${NC}    : $WARN"
echo ""

if [[ $FAIL -eq 0 ]]; then
  echo -e "${GREEN}${BOLD}🎉 All assertions passed — ProxyMaze looks solid!${NC}"
elif [[ $FAIL -le 5 ]]; then
  echo -e "${YELLOW}${BOLD}⚠  A few failures — review and fix before submission.${NC}"
else
  echo -e "${RED}${BOLD}✗  Multiple failures — significant issues to address.${NC}"
fi

echo ""
echo -e "Webhook log: ${CYAN}$WEBHOOK_LOG${NC}"
echo -e "Run:  ${CYAN}cat $WEBHOOK_LOG | python3 -c \"import sys,json; [print(json.dumps(json.loads(l),indent=2)) for l in sys.stdin]\"${NC}  to inspect payloads."

exit $FAIL
