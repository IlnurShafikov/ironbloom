// Package calculator provides ways-win calculation for game boards,
// including support for wild symbols that add extra ways.
package calculator

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/limiter"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// Calculators defines the win calculation operations used by the game phases.
// It allows the phases to be tested with a mock instead of a real evaluator.
//
//go:generate mockgen -package=calculator -destination=calculators.mock.go -source=calculators.go Calculators
type Calculators interface {
	// CalculateWays evaluates the board and returns all ways wins,
	// applying the win limiter to every payout.
	CalculateWays(
		board model.Board, winLimiter limiter.WinLimiter,
	) (calculators.WaysWins, error)

	// CalculateWaysWithMultipliers evaluates the board like CalculateWays,
	// but also takes wild multipliers into account when counting ways.
	CalculateWaysWithMultipliers(
		board model.Board,
		winLimiter limiter.WinLimiter,
		wilds []result.WildMultiplier,
	) (calculators.WaysWins, error)
}

// Calculator is the default Calculators implementation based on the
// slot-engine ways evaluator.
type Calculator struct {
	// ways is the underlying ways evaluator.
	ways *calculators.Ways
}

// NewCalculator creates a Calculator on top of the given ways evaluator.
//
// Parameters:
//   - ways: configured ways evaluator (paytable, wilds and skipped symbols)
//
// Returns:
//   - pointer to a new Calculator
func NewCalculator(ways *calculators.Ways) *Calculator {
	return &Calculator{
		ways: ways,
	}
}

// CalculateWays evaluates the board and returns all ways wins.
//
// Parameters:
//   - board: board to evaluate
//   - winLimiter: limiter that caps the total win
//
// Returns:
//   - list of ways wins found on the board
//   - error if the evaluation fails
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

// CalculateWaysWithMultipliers evaluates the board and returns all ways wins,
// where every wild multiplier adds extra ways to the winning combinations.
//
// Parameters:
//   - board: board to evaluate
//   - winLimiter: limiter that caps the total win
//   - wilds: wild symbols with their multipliers found on the board
//
// Returns:
//   - list of ways wins found on the board
//   - error if the evaluation fails
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
