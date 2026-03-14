#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
fail() { echo -e "${RED}✗${NC} $*"; exit 1; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }

# Preflight: require jq for E2E JSON parsing
if ! command -v jq &>/dev/null; then
    warn "jq not found — license E2E step will be skipped even if vars are set"
    JQ_AVAILABLE=0
else
    JQ_AVAILABLE=1
fi

echo "=== polytrade-bot pre-release check ==="
echo ""

# ── Step 1: go vet ────────────────────────────────────────────────────────────
echo "Step 1: go vet"
go vet ./... && ok "go vet passed" || fail "go vet failed"

# ── Step 2: go test ───────────────────────────────────────────────────────────
echo ""
echo "Step 2: go test"
go test ./... -count=1 -timeout 60s && ok "all tests passed" || fail "tests failed"

# ── Step 3: env var check ─────────────────────────────────────────────────────
echo ""
echo "Step 3: env vars"
if [ -z "${POLY_PRIVATE_KEY:-}" ]; then
    warn "POLY_PRIVATE_KEY not set — L1 integration tests will be skipped"
else
    ok "POLY_PRIVATE_KEY is set"
fi
# rawToken is embedded via ldflags at release build time — not an env var here

# ── Step 4: cross-platform build ──────────────────────────────────────────────
echo ""
echo "Step 4: cross-platform build"
mkdir -p dist

GOOS=linux   GOARCH=amd64  go build -o dist/polytrade-bot-linux-amd64       ./cmd/bot \
    && ok "linux/amd64   → dist/polytrade-bot-linux-amd64" \
    || fail "linux/amd64 build failed"

GOOS=darwin  GOARCH=arm64  go build -o dist/polytrade-bot-darwin-arm64      ./cmd/bot \
    && ok "darwin/arm64  → dist/polytrade-bot-darwin-arm64" \
    || fail "darwin/arm64 build failed"

GOOS=windows GOARCH=amd64  go build -o dist/polytrade-bot-windows-amd64.exe ./cmd/bot \
    && ok "windows/amd64 → dist/polytrade-bot-windows-amd64.exe" \
    || fail "windows/amd64 build failed"

# ── Step 5: license E2E (optional) ───────────────────────────────────────────
echo ""
echo "Step 5: license E2E"
if [ -z "${ORBITRON_LICENSE_URL:-}" ] || [ -z "${POLY_APP_TOKEN:-}" ] || [ "$JQ_AVAILABLE" = "0" ]; then
    warn "ORBITRON_LICENSE_URL, POLY_APP_TOKEN, or jq not available — E2E skipped"
    warn "To run: export ORBITRON_LICENSE_URL=... POLY_APP_TOKEN=... && bash scripts/check-bot.sh"
else
    HTTP_CODE=$(curl -s -o /tmp/lic_resp.json -w "%{http_code}" \
        -X POST "$ORBITRON_LICENSE_URL" \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$POLY_APP_TOKEN\",\"version\":\"check\"}" \
        --max-time 10)

    if [ "$HTTP_CODE" != "200" ]; then
        fail "license E2E: expected HTTP 200, got $HTTP_CODE (body: $(cat /tmp/lic_resp.json))"
    fi

    API_KEY=$(jq -r '.builder_api_key // empty' /tmp/lic_resp.json)
    if [ -z "$API_KEY" ]; then
        fail "license E2E: builder_api_key missing or empty in response"
    fi

    ok "license E2E passed (HTTP 200, builder_api_key present)"
fi

echo ""
ok "=== All checks passed. Ready to release. ==="
