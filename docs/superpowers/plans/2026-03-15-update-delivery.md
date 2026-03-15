# Update Delivery System — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver automatic update notifications and self-updates to all Orbitron users via a public version endpoint on getorbitron.net and an `internal/updater/` package in the bot.

**Architecture:** The website exposes `GET /api/v1/version` (env-driven, rate-limited to 60/hr). The bot checks it at startup and every 6 hours; if a newer version is found, it notifies via TUI banner + Telegram + log, then applies the update when idle (git pull + restart, or binary replacement).

**Tech Stack:** Next.js App Router (TypeScript) for the server; standard library `net/http`, `golang.org/x/mod/semver`, `os/exec`, `syscall` for the bot.

---

## Chunk 1: Website — GET /api/v1/version

### Task 1: Extend `checkRateLimit` to accept an optional `limit` parameter

**Files:**
- Modify: `src/lib/license-api.ts`
- Modify: `src/lib/license-api.test.ts`

The current `checkRateLimit` uses the module-level `RATE_LIMIT = 10` constant internally. Add an optional `limit` parameter so callers can pass a different cap. Existing call sites pass no `limit` argument and remain unchanged.

- [ ] **Step 1: Write failing test for new `limit` parameter**

  Add to `src/lib/license-api.test.ts`:
  ```ts
  it("respects custom limit of 2", () => {
    const map: RateMap = new Map();
    // now=0 represents the start of the rate window (arbitrary epoch)
    expect(checkRateLimit("ip", map, 0, 2)).toBe(true);  // 1st request
    expect(checkRateLimit("ip", map, 1, 2)).toBe(true);  // 2nd request
    expect(checkRateLimit("ip", map, 2, 2)).toBe(false); // 3rd → blocked
  });
  ```

- [ ] **Step 2: Run test — expect FAIL**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/TSJS/orbitron-polymarket-website"
  npx vitest run src/lib/license-api.test.ts
  ```
  Expected: fails because `checkRateLimit` ignores 4th argument.

- [ ] **Step 3: Add `limit` parameter to `checkRateLimit`**

  In `src/lib/license-api.ts`, change the signature and the comparison:
  ```ts
  export function checkRateLimit(
    ip: string,
    map: RateMap,
    now: number,
    limit: number = RATE_LIMIT,
  ): boolean {
    const entry = map.get(ip);
    if (!entry || now > entry.resetAt) {
      map.set(ip, { count: 1, resetAt: now + RATE_WINDOW_MS });
      return true;
    }
    if (entry.count >= limit) return false;
    entry.count++;
    return true;
  }
  ```

- [ ] **Step 4: Run all lib tests — expect PASS**

  ```bash
  npx vitest run src/lib/license-api.test.ts
  ```

- [ ] **Step 5: Commit**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/TSJS/orbitron-polymarket-website"
  git add src/lib/license-api.ts src/lib/license-api.test.ts
  git commit -m "feat(license-api): add optional limit param to checkRateLimit"
  ```

---

### Task 2: Create `GET /api/v1/version` route

**Files:**
- Create: `src/app/api/v1/version/route.ts`
- Create: `src/app/api/v1/version/route.test.ts`

- [ ] **Step 1: Write failing tests**

  Create `src/app/api/v1/version/route.test.ts`:
  ```ts
  import { describe, it, expect, beforeEach, vi } from "vitest";
  import { GET } from "./route";
  import { NextRequest } from "next/server";

  function makeReq(ip = "1.2.3.4"): NextRequest {
    return new NextRequest("http://localhost/api/v1/version", {
      headers: { "x-forwarded-for": ip },
    });
  }

  describe("GET /api/v1/version", () => {
    beforeEach(() => {
      vi.stubEnv("VERSION_LATEST", "1.1.0");
      vi.stubEnv("VERSION_RELEASE_NOTES", "Bug fixes");
      vi.stubEnv("VERSION_PUBLISHED_AT", "2026-03-15T12:00:00Z");
      vi.stubEnv("VERSION_BIN_LINUX_AMD64", "https://example.com/linux-amd64");
      vi.stubEnv("VERSION_BIN_LINUX_ARM64", "https://example.com/linux-arm64");
      vi.stubEnv("VERSION_BIN_DARWIN_AMD64", "https://example.com/darwin-amd64");
      vi.stubEnv("VERSION_BIN_DARWIN_ARM64", "https://example.com/darwin-arm64");
      vi.stubEnv("VERSION_BIN_WINDOWS_AMD64", "https://example.com/windows-amd64.exe");
    });

    it("returns 503 when VERSION_LATEST is not set", async () => {
      vi.stubEnv("VERSION_LATEST", "");
      const res = await GET(makeReq());
      expect(res.status).toBe(503);
    });

    it("returns 200 with correct shape when all env vars are set", async () => {
      const res = await GET(makeReq());
      expect(res.status).toBe(200);
      const body = await res.json();
      expect(body.version).toBe("1.1.0");
      expect(body.release_notes).toBe("Bug fixes");
      expect(body.published_at).toBe("2026-03-15T12:00:00Z");
      expect(body.binaries.linux_amd64).toBe("https://example.com/linux-amd64");
      expect(body.binaries.windows_amd64).toBe("https://example.com/windows-amd64.exe");
    });

    // NOTE: this test must run in isolation (separate vitest run) or use a
    // unique IP to avoid sharing the module-level rateMap with other tests.
    it("returns 429 after 60 requests from the same IP", async () => {
      const ip = "9.9.9.9";
      for (let i = 0; i < 60; i++) {
        const res = await GET(makeReq(ip));
        expect(res.status).toBe(200);
      }
      const res = await GET(makeReq(ip));
      expect(res.status).toBe(429);
    });
  });
  ```

- [ ] **Step 2: Run tests — expect FAIL (module not found)**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/TSJS/orbitron-polymarket-website"
  npx vitest run src/app/api/v1/version/route.test.ts
  ```

- [ ] **Step 3: Create the route**

  Create `src/app/api/v1/version/route.ts`:
  ```ts
  import { NextRequest, NextResponse } from "next/server";
  import { checkRateLimit, type RateMap } from "@/lib/license-api";

  const VERSION_RATE_LIMIT = 60;
  const rateMap: RateMap = new Map();

  export async function GET(req: NextRequest) {
    const ip =
      req.headers.get("x-forwarded-for")?.split(",")[0].trim() ?? "unknown";

    if (!checkRateLimit(ip, rateMap, Date.now(), VERSION_RATE_LIMIT)) {
      return NextResponse.json({ error: "rate_limited" }, { status: 429 });
    }

    const version = process.env.VERSION_LATEST ?? "";
    if (!version) {
      return NextResponse.json(
        { error: "version_not_configured" },
        { status: 503 },
      );
    }

    return NextResponse.json({
      version,
      release_notes: process.env.VERSION_RELEASE_NOTES ?? "",
      published_at:  process.env.VERSION_PUBLISHED_AT  ?? "",
      binaries: {
        linux_amd64:   process.env.VERSION_BIN_LINUX_AMD64   ?? "",
        linux_arm64:   process.env.VERSION_BIN_LINUX_ARM64   ?? "",
        darwin_amd64:  process.env.VERSION_BIN_DARWIN_AMD64  ?? "",
        darwin_arm64:  process.env.VERSION_BIN_DARWIN_ARM64  ?? "",
        windows_amd64: process.env.VERSION_BIN_WINDOWS_AMD64 ?? "",
      },
    });
  }
  ```

- [ ] **Step 4: Run tests — expect PASS**

  ```bash
  npx vitest run src/app/api/v1/version/route.test.ts
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add src/app/api/v1/version/route.ts src/app/api/v1/version/route.test.ts
  git commit -m "feat(api): add GET /api/v1/version endpoint"
  ```

---

## Chunk 2: Bot — internal/updater/

### Task 3: Add `IsIdle()` to `trading.Engine`

**Files:**
- Modify: `internal/trading/engine.go`
- Modify: `internal/trading/engine_test.go`

`IsIdle()` returns `true` when no strategy has a non-nil `cancel` (meaning no strategy is running via `StartStrategy`). The existing `fakeStrategy` in `engine_test.go` blocks until `ctx.Done()` — use it directly. `StartStrategy` sets `cancel` synchronously before launching the goroutine, so the `IsIdle()` check immediately after is race-free.

- [ ] **Step 1: Write failing test**

  Add to `internal/trading/engine_test.go`:
  ```go
  func TestEngine_IsIdle(t *testing.T) {
      log := zerolog.Nop()
      e := NewEngine(log)
      // No strategies registered → idle
      assert.True(t, e.IsIdle())

      ctx, cancel := context.WithCancel(context.Background())
      defer cancel()

      s := &fakeStrategy{name: "s1"} // blocks until ctx.Done()
      e.Register(s)
      // Registered but not started → idle
      assert.True(t, e.IsIdle())

      require.NoError(t, e.StartStrategy(ctx, "s1"))
      // cancel is set synchronously inside StartStrategy → not idle
      assert.False(t, e.IsIdle())

      require.NoError(t, e.StopStrategy("s1"))
      // cancel cleared → idle again
      assert.True(t, e.IsIdle())
  }
  ```

- [ ] **Step 2: Run test — expect FAIL**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
  go test ./internal/trading/... -run TestEngine_IsIdle -v
  ```

- [ ] **Step 3: Add `IsIdle()` to engine.go**

  Add after the `Stop()` method:
  ```go
  // IsIdle returns true when no strategy is currently running.
  func (e *Engine) IsIdle() bool {
      e.mu.RLock()
      defer e.mu.RUnlock()
      for _, en := range e.entries {
          if en.cancel != nil {
              return false
          }
      }
      return true
  }
  ```

- [ ] **Step 4: Run test — expect PASS**

  ```bash
  go test ./internal/trading/... -run TestEngine_IsIdle -v
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add internal/trading/engine.go internal/trading/engine_test.go
  git commit -m "feat(trading): add IsIdle() to Engine"
  ```

---

### Task 4: Add `UpdateAvailableMsg` to TUI messages

**Files:**
- Modify: `internal/tui/messages.go`

- [ ] **Step 1: Add the message type**

  Open `internal/tui/messages.go` and add:
  ```go
  // UpdateAvailableMsg is published to EventBus when a newer bot version is detected.
  type UpdateAvailableMsg struct {
      Version      string
      ReleaseNotes string
      PublishedAt  string
  }
  ```

- [ ] **Step 2: Verify compilation**

  ```bash
  go build ./internal/tui/...
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add internal/tui/messages.go
  git commit -m "feat(tui): add UpdateAvailableMsg type"
  ```

---

### Task 5: `internal/updater/pending.go`

**Files:**
- Create: `internal/updater/pending.go`
- Create: `internal/updater/pending_test.go`

- [ ] **Step 1: Write failing tests**

  Create `internal/updater/pending_test.go`:
  ```go
  package updater_test

  import (
      "os"
      "path/filepath"
      "testing"

      "github.com/atlasdev/orbitron/internal/updater"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestPending_RoundTrip(t *testing.T) {
      dir := t.TempDir()
      p := updater.NewPending(dir)

      _, ok := p.Load()
      assert.False(t, ok, "no file → Load returns false")

      p.Save("1.1.0", "https://example.com/bin")
      got, ok := p.Load()
      require.True(t, ok)
      assert.Equal(t, "1.1.0", got.Version)
      assert.Equal(t, "https://example.com/bin", got.BinaryURL)

      p.Clear()
      _, ok = p.Load()
      assert.False(t, ok, "after Clear → Load returns false")
  }

  func TestPending_MalformedFile(t *testing.T) {
      dir := t.TempDir()
      p := updater.NewPending(dir)

      err := os.WriteFile(filepath.Join(dir, ".update_pending"), []byte("not-json"), 0o600)
      require.NoError(t, err)

      _, ok := p.Load()
      assert.False(t, ok, "malformed JSON → Load returns false and deletes file")

      // File must be deleted after malformed read
      _, err = os.Stat(filepath.Join(dir, ".update_pending"))
      assert.True(t, os.IsNotExist(err))
  }
  ```

- [ ] **Step 2: Run — expect FAIL**

  ```bash
  go test ./internal/updater/... -run TestPending -v
  ```

- [ ] **Step 3: Create `pending.go`**

  Create `internal/updater/pending.go`:
  ```go
  package updater

  import (
      "encoding/json"
      "os"
      "path/filepath"

      "github.com/rs/zerolog/log"
  )

  const pendingFile = ".update_pending"

  // PendingUpdate holds the deferred update state written to disk.
  type PendingUpdate struct {
      Version   string `json:"version"`
      BinaryURL string `json:"binary_url"`
  }

  // Pending manages the .update_pending file in a given directory.
  type Pending struct {
      path string
  }

  // NewPending returns a Pending rooted at dir.
  func NewPending(dir string) *Pending {
      return &Pending{path: filepath.Join(dir, pendingFile)}
  }

  // Save writes version and binaryURL to the pending file.
  func (p *Pending) Save(version, binaryURL string) {
      data, _ := json.Marshal(PendingUpdate{Version: version, BinaryURL: binaryURL})
      if err := os.WriteFile(p.path, data, 0o600); err != nil {
          log.Warn().Err(err).Msg("updater: failed to save pending update")
      }
  }

  // Load reads the pending file. Returns (nil, false) when absent or malformed.
  func (p *Pending) Load() (*PendingUpdate, bool) {
      data, err := os.ReadFile(p.path)
      if os.IsNotExist(err) {
          return nil, false
      }
      var u PendingUpdate
      if err := json.Unmarshal(data, &u); err != nil || u.Version == "" {
          log.Warn().Msg("updater: malformed .update_pending — deleting")
          _ = os.Remove(p.path)
          return nil, false
      }
      return &u, true
  }

  // Clear deletes the pending file.
  func (p *Pending) Clear() {
      _ = os.Remove(p.path)
  }
  ```

- [ ] **Step 4: Run — expect PASS**

  ```bash
  go test ./internal/updater/... -run TestPending -v
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add internal/updater/pending.go internal/updater/pending_test.go
  git commit -m "feat(updater): add Pending read/write/clear"
  ```

---

### Task 6: `internal/updater/notifier.go`

**Files:**
- Create: `internal/updater/notifier.go`

No separate unit test — notifier is thin glue; behaviour is verified end-to-end when wiring in Task 9.

- [ ] **Step 1: Create `notifier.go`**

  Create `internal/updater/notifier.go`:
  ```go
  package updater

  import (
      "context"
      "fmt"

      "github.com/atlasdev/orbitron/internal/notify"
      "github.com/atlasdev/orbitron/internal/tui"
      "github.com/rs/zerolog/log"
  )

  // Notifier sends update-available events to all channels.
  type Notifier struct {
      bus      *tui.EventBus
      telegram notify.Notifier
  }

  // NewNotifier creates a Notifier. Pass nil for channels you don't need.
  func NewNotifier(bus *tui.EventBus, telegram notify.Notifier) *Notifier {
      return &Notifier{bus: bus, telegram: telegram}
  }

  // Notify fires an update-available event to TUI, Telegram, and log.
  func (n *Notifier) Notify(version, releaseNotes, publishedAt string) {
      log.Info().
          Str("latest", version).
          Str("published_at", publishedAt).
          Str("notes", releaseNotes).
          Msg("update available")

      if n.bus != nil {
          n.bus.Send(tui.UpdateAvailableMsg{
              Version:      version,
              ReleaseNotes: releaseNotes,
              PublishedAt:  publishedAt,
          })
      }

      if n.telegram != nil {
          msg := fmt.Sprintf(
              "Orbitron update available: v%s\n%s\nPublished: %s",
              version, releaseNotes, publishedAt,
          )
          if err := n.telegram.Send(context.Background(), msg); err != nil {
              log.Warn().Err(err).Msg("updater: telegram notification failed")
          }
      }
  }

  // NotifyError sends an error notification via Telegram and log.
  func (n *Notifier) NotifyError(msg string) {
      log.Error().Msg("updater: " + msg)
      if n.telegram != nil {
          _ = n.telegram.Send(context.Background(), "Orbitron updater error: "+msg)
      }
  }
  ```

- [ ] **Step 2: Verify compilation**

  ```bash
  go build ./internal/updater/...
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add internal/updater/notifier.go
  git commit -m "feat(updater): add Notifier (TUI + Telegram + log)"
  ```

---

### Task 7: `internal/updater/checker.go`

**Files:**
- Create: `internal/updater/checker.go`
- Create: `internal/updater/checker_os.go`
- Create: `internal/updater/checker_test.go`

- [ ] **Step 0: Verify `golang.org/x/mod` is in go.mod**

  ```bash
  grep "golang.org/x/mod" go.mod
  ```
  Expected: a line like `golang.org/x/mod v0.x.x`. If absent, run `go get golang.org/x/mod`.

- [ ] **Step 1: Write failing tests**

  Create `internal/updater/checker_test.go`:
  ```go
  package updater_test

  import (
      "context"
      "testing"

      "github.com/atlasdev/orbitron/internal/updater"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestRemoteIsNewer(t *testing.T) {
      cases := []struct {
          remote, local string
          wantNewer     bool
      }{
          {"1.1.0", "1.0.0", true},
          {"1.0.0", "1.0.0", false},
          {"1.0.0", "1.1.0", false},
          {"1.10.0", "1.9.0", true},   // 10 > 9
          {"2.0.0", "1.99.99", true},
          {"bad", "1.0.0", false},     // malformed remote → not newer
      }
      for _, c := range cases {
          got := updater.RemoteIsNewer(c.remote, c.local)
          assert.Equal(t, c.wantNewer, got, "remote=%s local=%s", c.remote, c.local)
      }
  }

  func TestScheduleUpdate_NotIdle_SavesPending(t *testing.T) {
      dir := t.TempDir()
      p := updater.NewPending(dir)
      notIdle := func() bool { return false }

      updater.ScheduleUpdateWith(context.Background(), notIdle, p, "1.1.0", "https://example.com/bin", nil)

      got, ok := p.Load()
      require.True(t, ok)
      assert.Equal(t, "1.1.0", got.Version)
  }
  ```

- [ ] **Step 2: Run — expect FAIL**

  ```bash
  go test ./internal/updater/... -run "TestRemoteIsNewer|TestScheduleUpdate" -v
  ```

- [ ] **Step 3: Create `checker.go`**

  Create `internal/updater/checker.go`:
  ```go
  package updater

  import (
      "context"
      "encoding/json"
      "net/http"
      "time"

      "github.com/atlasdev/orbitron/internal/license"
      "github.com/rs/zerolog/log"
      "golang.org/x/mod/semver"
  )

  // VersionServerURL is the endpoint to check for updates.
  // Override in tests or via ldflags.
  var VersionServerURL = "https://getorbitron.net/api/v1/version"

  type versionResponse struct {
      Version      string            `json:"version"`
      ReleaseNotes string            `json:"release_notes"`
      PublishedAt  string            `json:"published_at"`
      Binaries     map[string]string `json:"binaries"`
  }

  // RemoteIsNewer returns true when remote version string is strictly greater than local.
  // Both strings are plain semver without the "v" prefix (e.g. "1.1.0").
  func RemoteIsNewer(remote, local string) bool {
      r := "v" + remote
      l := "v" + local
      if !semver.IsValid(r) || !semver.IsValid(l) {
          return false
      }
      return semver.Compare(r, l) > 0
  }

  // Start checks for updates on startup (handling any pending update first), then
  // rechecks every 6 hours until ctx is cancelled.
  // dir is the source root / executable directory (used for .git detection and
  // binary replacement). Pass updater.Dir() from main.
  func Start(ctx context.Context, isIdle func() bool, n *Notifier, p *Pending) {
      dir := Dir()

      // Handle a deferred update from a previous run.
      if u, ok := p.Load(); ok {
          n.Notify(u.Version, "", "")
          RunUpdate(ctx, u.Version, u.BinaryURL, dir, p, n)
      }

      checkVersion(ctx, dir, isIdle, n, p)

      ticker := time.NewTicker(6 * time.Hour)
      defer ticker.Stop()
      for {
          select {
          case <-ctx.Done():
              return
          case <-ticker.C:
              checkVersion(ctx, dir, isIdle, n, p)
          }
      }
  }

  func checkVersion(ctx context.Context, dir string, isIdle func() bool, n *Notifier, p *Pending) {
      client := &http.Client{Timeout: 10 * time.Second}
      resp, err := client.Get(VersionServerURL)
      if err != nil {
          log.Warn().Err(err).Msg("updater: version check failed")
          return
      }
      defer resp.Body.Close()

      var vr versionResponse
      if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil || vr.Version == "" {
          log.Warn().Msg("updater: invalid version response")
          return
      }

      if !RemoteIsNewer(vr.Version, license.Version) {
          return
      }

      binaryURL := vr.Binaries[binaryKey()]
      n.Notify(vr.Version, vr.ReleaseNotes, vr.PublishedAt)
      ScheduleUpdateWith(ctx, isIdle, p, vr.Version, binaryURL, func() {
          RunUpdate(ctx, vr.Version, binaryURL, dir, p, n)
      })
  }

  // ScheduleUpdateWith decides whether to run the update now or defer it.
  // When idle, it waits 30 seconds (cancellable via ctx) then calls runFn.
  // When not idle, it saves the pending update for the next startup.
  // Pass nil for runFn to test only the save-pending path.
  func ScheduleUpdateWith(ctx context.Context, isIdle func() bool, p *Pending, version, binaryURL string, runFn func()) {
      if isIdle() {
          if runFn != nil {
              go func() {
                  select {
                  case <-time.After(30 * time.Second):
                      runFn()
                  case <-ctx.Done():
                  }
              }()
          }
      } else {
          p.Save(version, binaryURL)
      }
  }
  ```

- [ ] **Step 4: Create `checker_os.go`** (build-time OS/arch constants)

  Create `internal/updater/checker_os.go`:
  ```go
  package updater

  import "runtime"

  // buildOS and buildArch reflect the OS/arch the binary was compiled for.
  var (
      buildOS   = runtime.GOOS
      buildArch = runtime.GOARCH
  )

  func binaryKey() string {
      return buildOS + "_" + buildArch
  }
  ```

- [ ] **Step 5: Run tests — expect PASS**

  ```bash
  go test ./internal/updater/... -run "TestRemoteIsNewer|TestScheduleUpdate" -v
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/updater/checker.go internal/updater/checker_os.go internal/updater/checker_test.go
  git commit -m "feat(updater): add checker with semver comparison and schedule logic"
  ```

---

### Task 8: `internal/updater/updater.go`

**Files:**
- Create: `internal/updater/updater.go`
- Create: `internal/updater/dir.go`
- Create: `internal/updater/restart_unix.go`
- Create: `internal/updater/restart_windows.go`

The update logic differs by mode:
- **Git mode** (`<dir>/.git/` exists): `cd <dir> && git pull && ./setup.sh`, then restart.
- **Binary mode** (no `.git/`): download `binaryURL` to temp file, `chmod +x`, rename over current executable, restart. **Skipped on Windows** (binary auto-update not supported).

`RunUpdate` and `Start` both use `Dir()` (from `dir.go`) internally so callers don't need to resolve the executable path themselves.

- [ ] **Step 1: Create `dir.go`** (single source of truth for the working directory)

  Create `internal/updater/dir.go`:
  ```go
  package updater

  import (
      "os"
      "path/filepath"
  )

  // sourceDir is the canonical resolver for the bot's source/binary directory.
  // Override in tests: updater.sourceDir = func() string { return t.TempDir() }
  var sourceDir = func() string {
      exe, err := os.Executable()
      if err != nil {
          return "."
      }
      return filepath.Dir(exe)
  }

  // Dir returns the source/binary directory for this process.
  func Dir() string { return sourceDir() }
  ```

- [ ] **Step 2: Create `updater.go`**

  Create `internal/updater/updater.go`:
  ```go
  package updater

  import (
      "context"
      "fmt"
      "io"
      "net/http"
      "os"
      "os/exec"
      "path/filepath"
      "runtime"
      "time"

      "github.com/rs/zerolog/log"
  )

  // RunUpdate applies the update. It prefers git pull when a .git directory is
  // present, otherwise downloads the binary directly. On success it restarts
  // the process in-place. n is used to send error notifications on failure.
  func RunUpdate(ctx context.Context, version, binaryURL, dir string, p *Pending, n *Notifier) {
      if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
          runGitUpdate(dir, version, p, n)
          return
      }
      runBinaryUpdate(ctx, binaryURL, version, dir, p, n)
  }

  func runGitUpdate(dir, version string, p *Pending, n *Notifier) {
      log.Info().Str("version", version).Msg("updater: running git pull")

      pull := exec.Command("git", "pull")
      pull.Dir = dir
      if out, err := pull.CombinedOutput(); err != nil {
          msg := fmt.Sprintf("git pull failed: %v — %s", err, out)
          n.NotifyError(msg)
          return
      }

      setup := filepath.Join(dir, "setup.sh")
      if _, err := os.Stat(setup); err == nil {
          cmd := exec.Command("bash", setup)
          cmd.Dir = dir
          if out, err := cmd.CombinedOutput(); err != nil {
              msg := fmt.Sprintf("setup.sh failed: %v — %s", err, out)
              n.NotifyError(msg)
              return
          }
      }

      p.Clear()
      log.Info().Str("version", version).Msg("updater: git update complete, restarting")
      restartProcess()
  }

  func runBinaryUpdate(ctx context.Context, binaryURL, version, dir string, p *Pending, n *Notifier) {
      if runtime.GOOS == "windows" {
          log.Warn().Msg("updater: binary auto-update not supported on Windows; update manually")
          return
      }
      if binaryURL == "" {
          log.Warn().Msg("updater: no binary URL for this platform")
          return
      }

      exe, err := os.Executable()
      if err != nil {
          n.NotifyError("cannot determine executable path: " + err.Error())
          return
      }
      _ = dir // dir is used for .git detection only; binary lives at exe

      tmp := exe + ".new"
      if err := downloadFile(ctx, binaryURL, tmp); err != nil {
          n.NotifyError("binary download failed: " + err.Error())
          p.Save(version, binaryURL)
          return
      }

      if err := os.Chmod(tmp, 0o755); err != nil {
          n.NotifyError("chmod failed: " + err.Error())
          _ = os.Remove(tmp)
          return
      }

      if err := os.Rename(tmp, exe); err != nil {
          n.NotifyError("rename failed: " + err.Error())
          _ = os.Remove(tmp)
          p.Save(version, binaryURL)
          return
      }

      p.Clear()
      log.Info().Str("version", version).Msg("updater: binary replaced, restarting")
      restartProcess()
  }

  func downloadFile(ctx context.Context, url, dst string) error {
      client := &http.Client{Timeout: 5 * time.Minute}
      req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
      if err != nil {
          return fmt.Errorf("build request: %w", err)
      }
      resp, err := client.Do(req)
      if err != nil {
          return fmt.Errorf("download: %w", err)
      }
      defer resp.Body.Close()

      f, err := os.Create(dst)
      if err != nil {
          return fmt.Errorf("create temp file: %w", err)
      }
      defer f.Close()

      if _, err := io.Copy(f, resp.Body); err != nil {
          return fmt.Errorf("write: %w", err)
      }
      return nil
  }
  ```

- [ ] **Step 3: Create `restart_unix.go`** (syscall.Exec, Linux/macOS)

  Create `internal/updater/restart_unix.go`:
  ```go
  //go:build !windows

  package updater

  import (
      "os"
      "syscall"

      "github.com/rs/zerolog/log"
  )

  func restartProcess() {
      exe, err := os.Executable()
      if err != nil {
          log.Fatal().Err(err).Msg("updater: cannot get executable path for restart")
      }
      if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
          log.Fatal().Err(err).Msg("updater: exec restart failed")
      }
  }
  ```

- [ ] **Step 4: Create `restart_windows.go`**

  Create `internal/updater/restart_windows.go`:
  ```go
  //go:build windows

  package updater

  import "github.com/rs/zerolog/log"

  func restartProcess() {
      log.Warn().Msg("updater: restart not supported on Windows; please restart manually")
  }
  ```

- [ ] **Step 5: Verify compilation on current platform**

  ```bash
  go build ./internal/updater/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/updater/updater.go internal/updater/dir.go \
          internal/updater/restart_unix.go internal/updater/restart_windows.go
  git commit -m "feat(updater): implement RunUpdate (git pull + binary replacement + restart)"
  ```

---

### Task 9: Wire `updater.Start` in `cmd/bot/main.go`

**Files:**
- Modify: `cmd/bot/main.go`

`updater.Dir()` is the single source of truth for the working directory — do NOT call `os.Executable()` separately in main.go.

- [ ] **Step 1: Add imports and wire the updater**

  In `cmd/bot/main.go`, after all subsystems are initialised (trading engine, TUI EventBus, and Telegram notifier are all available):

  ```go
  import "github.com/atlasdev/orbitron/internal/updater"

  // After all subsystems are ready:
  pending := updater.NewPending(updater.Dir())
  updateNotifier := updater.NewNotifier(tuiBus, telegramNotifier) // pass nil if unavailable
  go updater.Start(ctx, tradingEngine.IsIdle, updateNotifier, pending)
  ```

- [ ] **Step 2: Verify full build**

  ```bash
  go build ./cmd/bot/...
  ```

- [ ] **Step 3: Run all unit tests**

  ```bash
  go test ./...
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add cmd/bot/main.go
  git commit -m "feat(main): wire updater.Start after subsystem initialization"
  ```

---

### Task 10: Wire `UpdateAvailableMsg` banner in TUI `app.go`

**Files:**
- Modify: `internal/tui/app.go`

The spec requires `app.go` to handle `UpdateAvailableMsg` and render a banner above the tab bar. Read `internal/tui/app.go` before making changes to follow the existing `Update`/`View` pattern.

- [ ] **Step 1: Read the existing app.go**

  ```bash
  # Identify the AppModel struct, Update() and View() methods
  grep -n "AppModel\|func.*Update\|func.*View\|banner\|Banner" internal/tui/app.go | head -30
  ```

- [ ] **Step 2: Add `updateBanner` field to AppModel**

  In the `AppModel` struct, add:
  ```go
  updateBanner string // non-empty when an update is available
  ```

- [ ] **Step 3: Handle `UpdateAvailableMsg` in `Update()`**

  In the `Update()` method's type switch, add (unqualified name — `app.go` is inside the `tui` package):
  ```go
  case UpdateAvailableMsg:
      m.updateBanner = fmt.Sprintf(
          " Update available: v%s — %s (published %s) ",
          msg.Version, msg.ReleaseNotes, msg.PublishedAt,
      )
  ```

- [ ] **Step 4: Render banner in `View()` above the tab bar**

  In the `View()` method, prepend the banner when non-empty. Use the existing `lipgloss` styles — a yellow/amber foreground consistent with `styles.go`:
  ```go
  var view string
  if m.updateBanner != "" {
      banner := lipgloss.NewStyle().
          Foreground(lipgloss.Color("#F5A623")).
          Bold(true).
          Render(m.updateBanner)
      view = banner + "\n"
  }
  view += // ... existing view assembly
  ```

- [ ] **Step 5: Verify compilation**

  ```bash
  go build ./internal/tui/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/tui/app.go
  git commit -m "feat(tui): render update-available banner above tab bar"
  ```

---

## Environment Variables to Add

After the website changes are deployed, set these in the production `.env` or Vercel dashboard. When releasing a new version, updating `VERSION_LATEST` (and related vars) in Vercel is the only deployment step — all running bots pick up the new version within 6 hours.

```
VERSION_LATEST=1.0.0
VERSION_RELEASE_NOTES=Initial release
VERSION_PUBLISHED_AT=2026-03-15T12:00:00Z
VERSION_BIN_LINUX_AMD64=https://github.com/atlasdev/orbitron/releases/download/v1.0.0/orbitron-linux-amd64
VERSION_BIN_LINUX_ARM64=https://github.com/atlasdev/orbitron/releases/download/v1.0.0/orbitron-linux-arm64
VERSION_BIN_DARWIN_AMD64=https://github.com/atlasdev/orbitron/releases/download/v1.0.0/orbitron-darwin-amd64
VERSION_BIN_DARWIN_ARM64=https://github.com/atlasdev/orbitron/releases/download/v1.0.0/orbitron-darwin-arm64
VERSION_BIN_WINDOWS_AMD64=https://github.com/atlasdev/orbitron/releases/download/v1.0.0/orbitron-windows-amd64.exe
```
