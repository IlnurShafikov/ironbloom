package events

import (
	"cmp"
	"slices"

	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// WinsEvent represents an event containing all winning lines and scatter wins from a spin.
// It includes all winning positions, line details with payouts, and scatter wins.
type WinsEvent struct {
	// Event is the event type identifier
	Event Type `json:"type"`
	// Positions contains all unique positions involved in wins (both line and scatter)
	Positions model.Positions `json:"positions"`
	// WaysWins contains all ways wins
	WaysWins []WaysWin `json:"ways,omitempty"`
}

func NewWinsEvent(
	waysWins calculators.WaysWins,
) *WinsEvent {
	waysUniquePos := collectUniqueWaysWinsPositions(waysWins)

	winWays := transferWays(waysWins, waysUniquePos)

	return &WinsEvent{
		Event:     WinsType,
		Positions: waysUniquePos,
		WaysWins:  winWays,
	}
}

// EventType returns the type identifier for this event.
func (c *WinsEvent) EventType() Type {
	return c.Event
}

type WaysWin struct {
	// PosIndices contains indices into the Positions array indicating winning positions
	PosIndices []int `json:"posIndices"`

	// Payout is the win amount for this scatter win
	Payout uint64 `json:"payout"`

	Ways int `json:"ways"`

	// Symbol is the symbol ID that formed the scatter win
	Symbol model.Symbol `json:"symbol"`
}

// transferWays converts scatter wins into the event format with position indices.
// It collects unique positions and maps them to indices for compact representation.
//
// Parameters:
//   - wins: scatter wins from the game engine
//
// Returns:
//   - slice of AnywhereWin with mapped position indices
func transferWays(
	wins calculators.WaysWins,
	uniquePositions model.Positions,
) []WaysWin {
	if len(wins) == 0 {
		return nil
	}

	positionIndices := buildPositionIndices(uniquePositions)

	result := make([]WaysWin, len(wins))
	for i, win := range wins {
		result[i] = WaysWin{
			Payout:     win.TotalPayout,
			Symbol:     win.Symbol,
			Ways:       win.Ways,
			PosIndices: mapToIndices(win.Positions, positionIndices),
		}
	}

	return result
}

// collectUniqueWaysWinsPositions extracts all unique positions from scatter wins.
// Positions are deduplicated and sorted by reel and row indices.
//
// Parameters:
//   - wins: scatter wins to extract positions from
//
// Returns:
//   - sorted slice of unique positions
func collectUniqueWaysWinsPositions(wins calculators.WaysWins) model.Positions {
	return collectPositions(func(yield func(model.Position)) {
		for _, win := range wins {
			for _, pos := range win.Positions {
				yield(pos)
			}
		}
	})
}

// collectPositions is a helper function that collects unique positions using an iterator.
// It deduplicates positions and returns them sorted by reel and row indices.
//
// Parameters:
//   - iter: iterator function that yields positions
//
// Returns:
//   - sorted slice of unique positions
func collectPositions(iter func(yield func(model.Position))) model.Positions {
	positionMap := make(map[model.Position]struct{})

	iter(func(pos model.Position) {
		positionMap[pos] = struct{}{}
	})

	positions := make(model.Positions, 0, len(positionMap))
	for pos := range positionMap {
		positions = append(positions, pos)
	}

	slices.SortFunc(positions, comparePositions)

	return positions
}

// buildPositionIndices creates a mapping from positions to their array indices.
// This is used to convert positions to compact index references.
//
// Parameters:
//   - positions: slice of positions to index
//
// Returns:
//   - map from position to its index in the slice
func buildPositionIndices(positions model.Positions) map[model.Position]int {
	indices := make(map[model.Position]int, len(positions))
	for i, pos := range positions {
		indices[pos] = i
	}

	return indices
}

// mapToIndices converts a slice of positions to their corresponding indices.
// Uses the position-to-index mapping to create a compact array of indices.
//
// Parameters:
//   - positions: positions to convert to indices
//   - positionIndices: mapping from position to index
//
// Returns:
//   - slice of indices corresponding to the input positions
func mapToIndices(positions model.Positions, positionIndices map[model.Position]int) []int {
	indices := make([]int, 0, len(positions))

	for _, pos := range positions {
		if index, ok := positionIndices[pos]; ok {
			indices = append(indices, index)
		}
	}

	return indices
}

// comparePositions is a comparison function for sorting positions.
// Sorts first by reel index, then by row index.
//
// Parameters:
//   - src: first position to compare
//   - target: second position to compare
//
// Returns:
//   - negative if src < target, 0 if equal, positive if src > target
func comparePositions(src, target model.Position) int {
	if src.ReelIndex != target.ReelIndex {
		return cmp.Compare(src.ReelIndex, target.ReelIndex)
	}

	return cmp.Compare(src.RowIndex, target.RowIndex)
}
