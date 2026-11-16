package calculator

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// WithWaysWinMultiplier returns a ways strategy that recalculates the payout
// of every win using the multipliers of the wild symbols.
// When there are no wilds, the win is returned unchanged.
//
// Parameters:
//   - wilds: wild symbols with their multipliers found on the board
//
// Returns:
//   - strategy to pass to the ways evaluator
func WithWaysWinMultiplier(
	wilds []result.WildMultiplier,
) calculators.WaysStrategy {
	if len(wilds) == 0 {
		return withEmptyWaysMultiplier
	}

	return withWaysWinMultiplier(wilds)
}

// withWaysWinMultiplier builds a strategy that multiplies the payout per way
// by the number of ways, counting every wild multiplier on each reel as
// additional ways.
//
// Parameters:
//   - wilds: wild symbols with their multipliers found on the board
//
// Returns:
//   - strategy that sets the total payout of a win
func withWaysWinMultiplier(
	wilds []result.WildMultiplier,
) calculators.WaysStrategy {
	return func(
		win calculators.WaysWin,
	) (calculators.WaysWin, bool) {
		var waysCount uint64 = 1

		tail := getLastReelIndex(win.Positions)

		for reel := 0; reel <= tail; reel++ {
			countSymbols := countSymbolOnReel(reel, win.Positions)
			countMultiplier := countWildsOnReel(reel, wilds)

			count := countSymbols + countMultiplier

			if count > 0 {
				waysCount *= count
			}
		}

		win.TotalPayout = waysCount * win.PayoutPerWay

		return win, false
	}
}

// getLastReelIndex returns the reel index of the last winning position,
// or 0 when there are no positions.
func getLastReelIndex(positions model.Positions) int {
	if len(positions) > 0 {
		return positions[len(positions)-1].ReelIndex
	}

	return 0
}

// withEmptyWaysMultiplier is a strategy that leaves the win unchanged.
// It is used when the board has no wild multipliers.
func withEmptyWaysMultiplier(
	win calculators.WaysWin,
) (calculators.WaysWin, bool) {
	return win, false
}

// countSymbolOnReel returns the number of winning positions on the given reel.
// The first three reels are counted from the left, the rest from the right.
//
// Parameters:
//   - reel: reel index
//   - positions: winning positions ordered by reel
//
// Returns:
//   - number of winning symbols on the reel
func countSymbolOnReel(reel int, positions model.Positions) uint64 {
	switch reel {
	//nolint: mnd
	case 0, 1, 2:
		return leftToRightCounter(reel, positions)
	default:
		return rightToLeftCounter(reel, positions)
	}
}

// leftToRightCounter counts positions on the reel scanning from the first
// position and stops as soon as a later reel is reached.
//
// Parameters:
//   - reelIndex: reel index
//   - positions: winning positions ordered by reel
//
// Returns:
//   - number of positions on the reel
func leftToRightCounter(
	reelIndex int, positions model.Positions,
) (count uint64) {
	for _, winPos := range positions {
		if winPos.ReelIndex == reelIndex {
			count++
		} else if winPos.ReelIndex > reelIndex {
			break
		}
	}

	return count
}

// rightToLeftCounter counts positions on the reel scanning from the last
// position and stops as soon as an earlier reel is reached.
//
// Parameters:
//   - reelIndex: reel index
//   - positions: winning positions ordered by reel
//
// Returns:
//   - number of positions on the reel
func rightToLeftCounter(
	reelIndex int, positions model.Positions,
) (count uint64) {
	for i := len(positions) - 1; i >= 0; i-- {
		winPos := positions[i]

		if winPos.ReelIndex == reelIndex {
			count++
		} else if winPos.ReelIndex < reelIndex {
			break
		}
	}

	return count
}

// countWildsOnReel returns the number of extra ways that the wild multipliers
// add on the given reel. A wild with multiplier N adds N-1 ways.
//
// Parameters:
//   - reelIndex: reel index
//   - wilds: wild symbols with their multipliers found on the board
//
// Returns:
//   - number of additional ways on the reel
func countWildsOnReel(reelIndex int, wilds []result.WildMultiplier) uint64 {
	var count uint64

	for _, wild := range wilds {
		if wild.Position().ReelIndex == reelIndex && wild.MultiplierValue() > 1 {
			count += (wild.MultiplierValue() - 1)
		}
	}

	return count
}
