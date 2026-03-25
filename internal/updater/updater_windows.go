//go:build windows

package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const helperName = "updater-helper.exe"

// RunUpdate applies the update on Windows.
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
		n.NotifyError(fmt.Sprintf("git pull failed: %v — %s", err, out))
		return
	}

	setup := filepath.Join(dir, "setup.ps1")
	if _, err := os.Stat(setup); err == nil {
		cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", setup)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			n.NotifyError(fmt.Sprintf("setup.ps1 failed: %v — %s", err, out))
			return
		}
	}

	p.Clear()
	log.Info().Str("version", version).Msg("updater: git update complete, restarting")
	restartProcess()
}

func runBinaryUpdate(ctx context.Context, binaryURL, version, dir string, p *Pending, n *Notifier) {
	if binaryURL == "" {
		log.Warn().Msg("updater: no binary URL for windows_amd64")
		return
	}

	exe, err := os.Executable()
	if err != nil {
		n.NotifyError("cannot determine executable path: " + err.Error())
		return
	}

	newExe := exe + ".new"
	if err := downloadFile(ctx, binaryURL, newExe); err != nil {
		n.NotifyError("binary download failed: " + err.Error())
		p.Save(version, binaryURL)
		return
	}

	helperPath := filepath.Join(filepath.Dir(exe), helperName)
	if _, err := os.Stat(helperPath); err != nil {
		n.NotifyError("updater-helper.exe not found alongside binary — cannot apply update on Windows: " + helperPath)
		_ = os.Remove(newExe)
		p.Save(version, binaryURL)
		return
	}

	// Spawn helper as a detached process. It waits 1.5s then swaps the binaries.
	cmd := exec.Command(helperPath, exe, newExe)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
	if err := cmd.Start(); err != nil {
		n.NotifyError("failed to launch updater-helper: " + err.Error())
		_ = os.Remove(newExe)
		p.Save(version, binaryURL)
		return
	}

	p.Clear()
	log.Info().Str("version", version).Msg("updater: helper launched, exiting for binary swap")
	time.Sleep(200 * time.Millisecond) // allow zerolog to flush before exit
	os.Exit(0)
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: unexpected status %d", resp.StatusCode)
	}
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
