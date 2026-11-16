package calculator

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/limiter"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

//go:generate mockgen -package=calculators -destination=calculators.mock.go -source=calculators.go Calculators
type Calculators interface {
	CalculateWays(
		board model.Board, winLimiter limiter.WinLimiter,
	) (calculators.WaysWins, error)

	CalculateWaysWithMultipliers(
		board model.Board,
		winLimiter limiter.WinLimiter,
		wilds []result.WildMultiplier,
	) (calculators.WaysWins, error)
}

type Calculator struct {
	ways *calculators.Ways
}

func NewCalculator(ways *calculators.Ways) *Calculator {
	return &Calculator{
		ways: ways,
	}
}

func (c *Calculator) CalculateWays(
	board model.Board, winLimiter limiter.WinLimiter,
) (calculators.WaysWins, error) {
	wins, err := c.ways.Calculate(board,
		calculators.WithWaysWinLimiter(winLimiter))
	if err != nil {
		return nil, fmt.Errorf("calculate ways: %w", err)
	}

	return wins, nil
}

func (c *Calculator) CalculateWaysWithMultipliers(
	board model.Board,
	winLimiter limiter.WinLimiter,
	wilds []result.WildMultiplier,
) (calculators.WaysWins, error) {
	wins, err := c.ways.Calculate(
		board,
		WithWaysWinMultiplier(wilds),
		calculators.WithWaysWinLimiter(winLimiter))
	if err != nil {
		return nil, fmt.Errorf("calculate ways: %w", err)
	}

	return wins, nil
}
