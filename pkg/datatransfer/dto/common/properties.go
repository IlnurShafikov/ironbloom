package common

// SymbolProperties contains optional properties that can be attached to symbols.
// These properties modify the behavior or display of symbols.
type SymbolProperties struct {
	// Multipier is the multiplier value of the symbol.
	Multipier uint64 `json:"multiplier"`
}
