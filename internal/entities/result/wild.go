package result

import "github.com/IlnurShafikov/slot-engine/pkg/model"

type WildMultiplier struct {
	position   model.Position
	multiplier uint64
	symbolID   model.Symbol
}

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

func (w WildMultiplier) MultiplierValue() uint64 {
	return w.multiplier
}

func (w WildMultiplier) Position() model.Position {
	return w.position
}

func (w WildMultiplier) SymbolID() model.Symbol {
	return w.symbolID
}
