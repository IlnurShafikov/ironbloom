// Package base implements the base game phase logic.
// The base phase represents the regular paid spins in the game.
package base

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/calculator"
	"github.com/IlnurShafikov/ironbloom/internal/components/collectors"
	"github.com/IlnurShafikov/ironbloom/internal/components/limiter"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/internal/tools/cast"
	"github.com/IlnurShafikov/slot-engine/pkg/generator"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/fscounter"
	"github.com/IlnurShafikov/slot-engine/pkg/weight"
)

// SpinPhase represents the base game spin phase with all its components.
// It contains the game mechanics and configurations needed to execute a base spin.
type SpinPhase struct {
	// gen produces raw reel boards for a given group ID.
	gen *generator.Generator

	// dist define the weighted symbol distributions for reels
	dist *weight.Weights

	// calc evaluates ways wins on the board.
	calc calculator.Calculators

	// fSInitCounter determines initial free spins count
	fSInitCounter fscounter.FSInitCounter
}

// NewSpinPhase creates a new base game spin phase with the specified components.
// Returns:
//   - pointer to a new SpinPhase instance
func NewSpinPhase(
	dist *weight.Weights,
	gen *generator.Generator,
	calc calculator.Calculators,
	fSInitCounter fscounter.FSInitCounter,
) *SpinPhase {
	return &SpinPhase{
		dist:          dist,
		gen:           gen,
		calc:          calc,
		fSInitCounter: fSInitCounter,
	}
}

// Spin executes the base game phase and returns the result.
// This is where all base game mechanics are implemented: reel spins, win calculation,
// feature triggers, etc.
//
// Parameters:
//   - rng: random number generator for deterministic randomness
//
// Returns:
//   - pointer to Base containing the spin result
//   - error if spin execution fails
func (b *SpinPhase) Spin(
	rng model.RNG,
	winLimiter limiter.WinLimiter,
) (*result.Base, error) {
	eventer := collectors.NewEvents()

	groupID, err := b.dist.GetRandomValue(rng)
	if err != nil {
		return nil, fmt.Errorf("distributions get random value: %w", err)
	}

	board, err := b.gen.Generate(rng, groupID)
	if err != nil {
		return nil, fmt.Errorf("generate board: %w", err)
	}

	totalWin, err := b.processWin(board, eventer, winLimiter)
	if err != nil {
		return nil, fmt.Errorf("process win: %w", err)
	}

	res := result.NewBase(board)

	err = b.checkBonusTrigger(res)
	if err != nil {
		return nil, fmt.Errorf("check bonus trigger: %w", err)
	}

	res.SetTotalWin(totalWin)
	res.SetMaxWinReached(winLimiter.MaxWinReached())
	res.SetEvents(eventer.Get())

	return res, nil
}

// processWin evaluates ways wins and coin wins for the generated board result,
// records corresponding events, and returns the combined payout.
// When the win limiter is hit, it returns the capped total immediately.
//
// Parameters:
//   - resp: board result containing the symbol grid and collector state
//   - eventer: event collector that accumulates game events for client delivery
//   - limiter: win cap enforcing the maximum allowed payout
//
// Returns:
//   - combined total of ways wins , capped by the limiter if reached
//   - error if ways calculation or event creation fails
func (b *SpinPhase) processWin(
	board model.Board,
	eventer *collectors.EventsCollector,
	winLimiter limiter.WinLimiter,
) (uint64, error) {
	waysWin, err := b.calc.CalculateWays(board, winLimiter)
	if err != nil {
		return 0, fmt.Errorf("calculate win lines: %w", err)
	}

	if len(waysWin) != 0 {
		eventer.Add(events.NewWinsEvent(waysWin))
	}

	return winLimiter.TotalWin(), nil
}

// checkBonusTrigger sets the number of free spins awarded by scatter symbols
// on the board. Nothing is changed when the scatters do not trigger the bonus.
//
// Parameters:
//   - res: result of the current base spin
//
// Returns:
//   - error if the free spins count cannot be converted to uint32
func (b *SpinPhase) checkBonusTrigger(
	res *result.Base,
) error {
	scatterCount := res.Board.CalculateSymbol(rules.ScatterID)
	fsCount := b.fSInitCounter.GetFsInitCount(scatterCount)

	if fsCount == 0 {
		return nil
	}

	safeValue, err := cast.SafeCast[uint32](fsCount)
	if err != nil {
		return fmt.Errorf("safe cast: %w", err)
	}

	res.SetFreeSpinsInitialsCount(safeValue)

	return nil
}
