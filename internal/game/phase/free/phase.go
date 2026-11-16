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

type SpinPhase struct {
	generator     *generator.Generator
	distributions *weight.Weights
	multiplierGen wild.WildsMultiplierGetter
	calculators   calculator.Calculators
	fSInitCounter fscounter.FSInitCounter
}

func NewSpinPhase(
	generator *generator.Generator,
	distributions *weight.Weights,
	multiplierGen wild.WildsMultiplierGetter,
	calculators calculator.Calculators,
	fSInitCounter fscounter.FSInitCounter,
) *SpinPhase {
	return &SpinPhase{
		generator:     generator,
		distributions: distributions,
		multiplierGen: multiplierGen,
		calculators:   calculators,
		fSInitCounter: fSInitCounter,
	}
}

func (f *SpinPhase) Spins(
	rng model.RNG,
	fsCount uint32,
	winLimiter limiter.WinLimiter,
) ([]*result.Free, error) {
	allRes := make([]*result.Free, 0, fsCount*freeSpinsCapacityFactor)

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

	err = f.checkRetrigger(res)
	if err != nil {
		return nil, fmt.Errorf("check retrigger: %w", err)
	}

	res.SetStepWin(stepWin)
	res.SetTotalWin(winLimiter.TotalWin())
	res.SetMaxWinReached(winLimiter.MaxWinReached())
	res.SetEvents(eventer.Get())

	return res, nil
}

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

func (f *SpinPhase) checkRetrigger(res *result.Free) error {
	scatterCount := res.Board.CalculateSymbol(rules.ScatterID)
	fsCount := f.fSInitCounter.GetFsInitCount(scatterCount)

	safeFsValue, err := cast.SafeCast[uint32](fsCount)
	if err != nil {
		return fmt.Errorf("safe cast: %w", err)
	}

	res.SetReTriggers(safeFsValue)

	return nil
}

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
