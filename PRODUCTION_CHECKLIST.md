# Production Checklist

Run these checks before every release. All boxes must be checked before proceeding.

---

## Bot Release (polytrade-bot → GitHub Release)

### Pre-release
- [ ] `bash scripts/check-bot.sh` completes with no ✗ errors
- [ ] All binaries in `dist/` are non-zero size:
  ```bash
  ls -lh dist/polytrade-bot-*
  ```
- [ ] `CHANGELOG.md` updated with the new version and release notes
- [ ] Git tag created:
  ```bash
  git tag v1.2.3
  git push origin v1.2.3
  ```
- [ ] GitHub Release created with all three binaries attached + install script

### Post-release verification
- [ ] Download URL works: `curl -L <release-asset-url> -o /tmp/test-bot`
- [ ] License E2E passed (optional — set `ORBITRON_LICENSE_URL` + `POLY_APP_TOKEN` and re-run `check-bot.sh`)

---

## Frontend Deploy (orbitron-polymarket-website → Ubuntu VPS)

### Pre-deploy
- [ ] `bash scripts/check-frontend.sh` completes with no ✗ errors
  - Requires: `APP_TOKEN`, `BUILDER_API_KEY`, `BUILDER_SECRET` exported
  - Optional smoke test: set `ORBITRON_BASE_URL` to the VPS URL before running
- [ ] VPS environment variables are current:
  ```bash
  ssh user@vps "grep -E 'APP_TOKEN|BUILDER' /path/to/.env"
  ```
- [ ] `git push origin main` → GitHub Actions deploy workflow triggered
- [ ] GitHub Actions workflow completed with green status

### Post-deploy verification
- [ ] Smoke test manually:
  ```bash
  curl -s -X POST https://your-vps-domain.com/api/v1/license \
    -H "Content-Type: application/json" \
    -d '{"token":"<APP_TOKEN>","version":"smoke"}' | jq .
  ```
  Expected: HTTP 200 + `builder_api_key` present in JSON response

---

## Environment Variables Reference

| Variable | Used by | Where to set | Notes |
|---|---|---|---|
| `APP_TOKEN` | Frontend (validates bot identity) | VPS `.env`, GitHub Secret | Must match token embedded in bot binary |
| `BUILDER_API_KEY` | Frontend (returned to bot) | VPS `.env`, GitHub Secret | Polymarket Builder Program key |
| `BUILDER_SECRET` | Frontend (returned to bot) | VPS `.env`, GitHub Secret | Polymarket Builder Program secret |
| `POLY_PRIVATE_KEY` | Bot (L1 EIP-712 signing) | User's `config.toml` or env | Optional — needed for trading only |
| `POLY_APP_TOKEN` | `check-bot.sh` E2E only | Local `.env` (not committed) | Plaintext token for E2E test |
| `ORBITRON_LICENSE_URL` | `check-bot.sh` E2E only | Local `.env` | Full URL of license endpoint |
| `ORBITRON_BASE_URL` | `check-frontend.sh` smoke | Local `.env` | Base URL of deployed frontend |
| `rawToken` (ldflags) | Bot binary | GitHub Actions release workflow | Set via `-ldflags "-X 'github.com/atlasdev/orbitron/internal/license.rawToken=ENCODED'"` |

### How to encode rawToken for ldflags

```bash
# In polytrade-bot repo:
go run ./cmd/tokenenc encode YOUR_PLAIN_TOKEN
# Copy the hex output → use as the ldflags value
```

---

## Rollback

### Bot rollback
Re-publish the previous GitHub Release tag. Users re-run the install script:
```bash
curl -fsSL https://your-install-url/install.sh | bash
```
The install script always pulls the latest release — pointing users to re-run with the
previous tag download URL is sufficient for immediate mitigation.

### Frontend rollback
```bash
git revert HEAD
git push origin main
# GitHub Actions automatically redeploys the reverted commit
```
Verify rollback with the smoke test above.
