// Package free implements the free spins phase logic.
// The free phase runs the bonus spins awarded by the base phase and by retriggers.
package free

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/calculator"
	"github.com/IlnurShafikov/ironbloom/internal/components/collectors"
	"github.com/IlnurShafikov/ironbloom/internal/components/limiter"
	"github.com/IlnurShafikov/ironbloom/internal/components/wild"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/internal/tools/cast"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/generator"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/fscounter"
	"github.com/IlnurShafikov/slot-engine/pkg/weight"
)

// freeSpinsCapacityFactor pre-sizes the free spins slice to hold the initial
// spins plus a typical number of retriggers.
const freeSpinsCapacityFactor = 2

// SpinPhase represents the free spins phase with all its components.
// It contains the game mechanics and configurations needed to execute free spins.
type SpinPhase struct {
	// generator produces raw reel boards for a given group ID.
	generator *generator.Generator

	// distributions define the weighted reel groups used for free spins.
	distributions *weight.Weights

	// multiplierGen finds wild symbols on the board and assigns their multipliers.
	multiplierGen wild.WildsMultiplierGetter

	// calculators evaluates ways wins on the board.
	calculators calculator.Calculators

	// fSInitCounter determines the number of free spins awarded for scatters.
	fSInitCounter fscounter.FSInitCounter
}

// NewSpinPhase creates a new free spins phase with the specified components.
//
// Parameters:
//   - gen: board generator for free spins
//   - distributions: weighted reel groups used for free spins
//   - multiplierGen: wild multiplier getter
//   - calc: ways wins calculator
//   - fSInitCounter: scatter count to free spins count mapping
//
// Returns:
//   - pointer to a new SpinPhase instance
func NewSpinPhase(
	gen *generator.Generator,
	distributions *weight.Weights,
	multiplierGen wild.WildsMultiplierGetter,
	calc calculator.Calculators,
	fSInitCounter fscounter.FSInitCounter,
) *SpinPhase {
	return &SpinPhase{
		generator:     gen,
		distributions: distributions,
		multiplierGen: multiplierGen,
		calculators:   calc,
		fSInitCounter: fSInitCounter,
	}
}

// Spins executes all free spins, including the ones added by retriggers.
// The phase stops early when the max win is reached.
//
// Parameters:
//   - rng: random number generator
//   - fsCount: initial number of free spins
//   - winLimiter: limiter that tracks the total win and the max win limit
//
// Returns:
//   - results of every executed free spin
//   - error if any spin fails
func (f *SpinPhase) Spins(
	rng model.RNG,
	fsCount uint32,
	winLimiter limiter.WinLimiter,
) ([]*result.Free, error) {
	allRes := make([]*result.Free, 0, fsCount*freeSpinsCapacityFactor)

	//nolint:intrange // fsCount grows on retriggers, so the bound is re-evaluated every iteration
	for fsIndex := 0; fsIndex < int(fsCount); fsIndex++ {
		fsResult, err := f.spin(rng, winLimiter)
		if err != nil {
			return nil, fmt.Errorf("spin %d failed: %w", fsIndex, err)
		}

		fsCount += fsResult.ReTriggers()

		allRes = append(allRes, fsResult)

		if winLimiter.MaxWinReached() {
			break
		}
	}

	return allRes, nil
}

// spin executes a single free spin: it picks a reel group, generates the board,
// finds wild multipliers, calculates wins and checks for a retrigger.
//
// Parameters:
//   - rng: random number generator
//   - winLimiter: limiter that tracks the total win and the max win limit
//
// Returns:
//   - result of the spin
//   - error if any step of the spin fails
func (f *SpinPhase) spin(
	rng model.RNG,
	winLimiter limiter.WinLimiter,
) (*result.Free, error) {
	eventer := collectors.NewEvents()

	groupID, err := f.distributions.GetRandomValue(rng)
	if err != nil {
		return nil, fmt.Errorf("get group id: %w", err)
	}

	board, err := f.generator.Generate(rng, groupID)
	if err != nil {
		return nil, fmt.Errorf("generate board: %w", err)
	}

	wildMultipliers, err := f.multiplierGen.GenerateWildsMultiplier(rng, board)
	if err != nil {
		return nil, fmt.Errorf("generate wild multiplier: %w", err)
	}

	res := result.NewFree(board, wildMultipliers)

	stepWin, err := f.processWin(res, eventer, winLimiter)
	if err != nil {
		return nil, fmt.Errorf("process line win: %w", err)
	}

	err = f.checkRetrigger(res, winLimiter)
	if err != nil {
		return nil, fmt.Errorf("check retrigger: %w", err)
	}

	res.SetStepWin(stepWin)
	res.SetTotalWin(winLimiter.TotalWin())
	res.SetMaxWinReached(winLimiter.MaxWinReached())
	res.SetEvents(eventer.Get())

	return res, nil
}

// processWin calculates ways wins of the spin taking wild multipliers into
// account and records the wins event.
//
// Parameters:
//   - res: result of the current spin
//   - eventer: collector of the spin events
//   - winLimiter: limiter that tracks the total win and the max win limit
//
// Returns:
//   - total win of the spin
//   - error if the calculation fails
func (f *SpinPhase) processWin(
	res *result.Free,
	eventer *collectors.EventsCollector,
	winLimiter limiter.WinLimiter,
) (uint64, error) {
	waysWins, err := f.calculators.CalculateWaysWithMultipliers(
		res.Board, winLimiter, res.Multipliers)
	if err != nil {
		return 0, fmt.Errorf("calculate ways wins: %w", err)
	}

	if len(waysWins) > 0 {
		eventer.Add(events.NewWinsEvent(waysWins))
	}

	stepWin := stepWinInWaysWins(waysWins)

	return stepWin, nil
}

// checkRetrigger sets the number of additional free spins awarded by scatter
// symbols on the board. Nothing is awarded when the max win is reached.
//
// Parameters:
//   - res: result of the current spin
//   - winLimiter: limiter that tracks the max win limit
//
// Returns:
//   - error if the free spins count cannot be converted to uint32
func (f *SpinPhase) checkRetrigger(res *result.Free, winLimiter limiter.WinLimiter) error {
	if winLimiter.MaxWinReached() {
		return nil
	}

	scatterCount := res.Board.CalculateSymbol(rules.ScatterID)
	fsCount := f.fSInitCounter.GetFsInitCount(scatterCount)

	safeFsValue, err := cast.SafeCast[uint32](fsCount)
	if err != nil {
		return fmt.Errorf("safe cast: %w", err)
	}

	res.SetReTriggers(safeFsValue)

	return nil
}

// stepWinInWaysWins sums the total payouts of all ways wins.
//
// Parameters:
//   - wins: ways wins of the spin
//
// Returns:
//   - sum of the payouts, 0 when there are no wins
func stepWinInWaysWins(wins calculators.WaysWins) uint64 {
	if len(wins) == 0 {
		return 0
	}

	var totalWin uint64

	for _, win := range wins {
		totalWin += win.TotalPayout
	}

	return totalWin
}
