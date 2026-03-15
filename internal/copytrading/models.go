// Package copytrading реализует автоматическое копирование сделок трейдеров Polymarket.
package copytrading

import (
	"github.com/atlasdev/orbitron/internal/api/data"
)

// TraderState — снимок позиций трейдера: map[assetID]Position.
type TraderState map[string]data.Position

// PositionDiff — результат сравнения двух снимков позиций.
type PositionDiff struct {
	// Opened — новые позиции (появились в текущем снимке)
	Opened []data.Position
	// Closed — закрытые позиции (исчезли из текущего снимка)
	Closed []data.Position
}

// diffStates сравнивает два снимка и возвращает изменения.
func diffStates(prev, curr TraderState) PositionDiff {
	var diff PositionDiff

	for assetID, pos := range curr {
		if _, exists := prev[assetID]; !exists {
			diff.Opened = append(diff.Opened, pos)
		}
	}

	for assetID, pos := range prev {
		if _, exists := curr[assetID]; !exists {
			diff.Closed = append(diff.Closed, pos)
		}
	}

	return diff
}

// toTraderState конвертирует список позиций в map по assetID.
func toTraderState(positions []data.Position) TraderState {
	state := make(TraderState, len(positions))
	for _, p := range positions {
		state[p.Asset] = p
	}
	return state
}
