// Package wild provides the generation of wild symbol multipliers.
package wild

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/entities/pairs"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/tools/cast"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// WildsMultiplierGetter defines how multipliers of wild symbols are found on a board.
//
//go:generate mockgen -package wild -destination=multiplier.mock.go -source=multiplier_getter.go WildsMultiplierGetter
type WildsMultiplierGetter interface {
	// GenerateWildsMultiplier returns every wild symbol on the board together
	// with its multiplier.
	GenerateWildsMultiplier(
		rng model.RNG,
		board model.Board,
	) ([]result.WildMultiplier, error)
}

var _ WildsMultiplierGetter = (*MultiplierGetter)(nil)

// MultiplierGetter finds wild symbols on a board and assigns them the
// multiplier configured for their symbol ID.
type MultiplierGetter struct {
	// wildPairMultiplier maps a wild symbol ID to its multiplier.
	wildPairMultiplier pairs.Pairs
	// wildIDs lists the symbol IDs that act as wilds.
	wildIDs []int
}

// NewMultiplierGetter creates a MultiplierGetter.
//
// Parameters:
//   - wildIDs: symbol IDs that act as wilds
//   - wildPairMultiplier: mapping of wild symbol ID to its multiplier
//
// Returns:
//   - pointer to a new MultiplierGetter
func NewMultiplierGetter(
	wildIDs []int,
	wildPairMultiplier pairs.Pairs,
) *MultiplierGetter {
	return &MultiplierGetter{
		wildIDs:            wildIDs,
		wildPairMultiplier: wildPairMultiplier,
	}
}

// GenerateWildsMultiplier finds all wild symbols on the board and returns
// them with the multiplier configured for their symbol ID.
//
// Parameters:
//   - rng: random number generator (not used by the current implementation)
//   - board: board to search for wild symbols
//
// Returns:
//   - list of wild symbols with their multipliers
//   - error if a wild has no configured multiplier or the multiplier does not fit uint64
func (m *MultiplierGetter) GenerateWildsMultiplier(
	rng model.RNG,
	board model.Board,
) ([]result.WildMultiplier, error) {
	var resp []result.WildMultiplier

	for _, wildID := range m.wildIDs {
		positions := board.SymbolPositions(wildID)

		for _, pos := range positions {
			safeMultiplier, err := m.multiplierByID(wildID)
			if err != nil {
				return nil, err
			}

			resp = append(resp, result.NewWildMultiplier(pos, safeMultiplier, wildID))
		}
	}

	return resp, nil
}

// multiplierByID returns the configured multiplier of the wild symbol.
//
// Parameters:
//   - wildID: wild symbol ID
//
// Returns:
//   - multiplier value
//   - error if the wild has no configured multiplier or the value does not fit uint64
func (m *MultiplierGetter) multiplierByID(wildID int) (uint64, error) {
	multiplier, ok := m.wildPairMultiplier.GetValueByKey(wildID)
	if !ok {
		return 0, ErrInvalidWildID
	}

	safeMultiplier, err := cast.SafeCast[uint64](multiplier)
	if err != nil {
		return 0, fmt.Errorf("safe int to uint64: %w", err)
	}

	return safeMultiplier, nil
}
