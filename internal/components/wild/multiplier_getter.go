package wild

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/entities/pairs"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/tools/cast"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

//go:generate mockgen -package wild -destination=multiplier.mock.go -source=multiplier_getter.go WildsMultiplierGetter
type WildsMultiplierGetter interface {
	GenerateWildsMultiplier(
		rng model.RNG,
		board model.Board,
	) ([]*result.WildMultiplier, error)
}

var _ WildsMultiplierGetter = (*MultiplierGetter)(nil)

type MultiplierGetter struct {
	wildIDs            []int
	wildPairMultiplier pairs.Pairs
}

func NewMultiplierGetter(
	wildIDs []int,
	wildPairMultiplier pairs.Pairs,
) *MultiplierGetter {
	return &MultiplierGetter{
		wildIDs:            wildIDs,
		wildPairMultiplier: wildPairMultiplier,
	}
}

func (m *MultiplierGetter) GenerateWildsMultiplier(
	rng model.RNG,
	board model.Board,
) ([]*result.WildMultiplier, error) {
	var resp []*result.WildMultiplier

	for _, wildID := range m.wildIDs {
		positions := board.SymbolPositions(wildID)
		if len(positions) == 0 {
			continue
		}

		for _, pos := range positions {
			multiplier, ok := m.wildPairMultiplier.GetValueByKey(wildID)
			if !ok {
				return nil, ErrInvalidWildID
			}

			safeMultiplier, err := cast.SafeCast[uint64](multiplier)
			if err != nil {
				return nil, fmt.Errorf("safe int to uint64: %w", err)
			}

			resp = append(resp, result.NewWildMultiplier(pos, safeMultiplier, wildID))
		}
	}

	return resp, nil
}
