//go:build windows

package updater

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

func restartProcess() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("updater: cannot get executable path for restart")
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		log.Fatal().Err(err).Msg("updater: restart failed")
	}
	os.Exit(0)
}
