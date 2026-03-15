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
