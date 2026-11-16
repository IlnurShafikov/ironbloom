package calculator

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

func WithWaysWinMultiplier(
	wilds []result.WildMultiplier,
) calculators.WaysStrategy {
	if len(wilds) == 0 {
		return withEmptyWaysMultiplier
	}

	return withWaysWinMultiplier(wilds)
}

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

func getLastReelIndex(positions model.Positions) int {
	if len(positions) > 0 {
		return positions[len(positions)-1].ReelIndex
	}

	return 0
}

func withEmptyWaysMultiplier(
	win calculators.WaysWin,
) (calculators.WaysWin, bool) {
	return win, false
}

func countSymbolOnReel(reel int, positions model.Positions) uint64 {
	switch reel {
	//nolint: mnd
	case 0, 1, 2:
		return leftToRightCounter(reel, positions)
	default:
		return rightToLeftCounter(reel, positions)
	}
}

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

func countWildsOnReel(reelIndex int, wilds []result.WildMultiplier) uint64 {
	var count uint64

	for _, wild := range wilds {
		if wild.Position().ReelIndex == reelIndex && wild.MultiplierValue() > 1 {
			count += (wild.MultiplierValue() - 1)
		}
	}

	return count
}
