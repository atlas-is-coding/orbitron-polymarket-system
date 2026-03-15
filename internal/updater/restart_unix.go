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
