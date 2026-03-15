//go:build windows

package updater

import "github.com/rs/zerolog/log"

func restartProcess() {
	log.Warn().Msg("updater: restart not supported on Windows; please restart manually")
}
