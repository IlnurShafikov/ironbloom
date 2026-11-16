package result

import "github.com/IlnurShafikov/slot-engine/pkg/model"

// WildMultiplier describes a wild symbol on the board and its multiplier.
type WildMultiplier struct {
	// position is the place of the wild on the board.
	position model.Position
	// multiplier is the multiplier value of the wild.
	multiplier uint64
	// symbolID is the ID of the wild symbol.
	symbolID model.Symbol
}

// NewWildMultiplier creates a WildMultiplier.
//
// Parameters:
//   - position: place of the wild on the board
//   - multiplier: multiplier value of the wild
//   - symbolID: ID of the wild symbol
//
// Returns:
//   - new WildMultiplier
func NewWildMultiplier(
	position model.Position,
	multiplier uint64,
	symbolID model.Symbol,
) WildMultiplier {
	return WildMultiplier{
		position:   position,
		multiplier: multiplier,
		symbolID:   symbolID,
	}
}

// MultiplierValue returns the multiplier value of the wild.
func (w WildMultiplier) MultiplierValue() uint64 {
	return w.multiplier
}

// Position returns the place of the wild on the board.
func (w WildMultiplier) Position() model.Position {
	return w.position
}

// SymbolID returns the ID of the wild symbol.
func (w WildMultiplier) SymbolID() model.Symbol {
	return w.symbolID
}
