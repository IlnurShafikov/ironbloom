// Package parser provides the RoundParser which processes raw game round results
// and accumulates them into a simulation Result for subsequent statistical analysis.
package parser

import (
	resp "github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/simulator/result"
)

// RoundParser processes round results and accumulates simulation statistics
type RoundParser struct {
	// res accumulates the statistics of all processed rounds.
	res *result.Result
}

// NewRoundParser creates a new RoundParser with initialized result container
func NewRoundParser(roundCost uint64, simpleCost int) *RoundParser {
	return &RoundParser{
		res: result.New(roundCost, simpleCost),
	}
}

// MergeResult merges a worker's accumulated result into this parser's result.
// Used by the parallel runner after all workers complete.
func (rp *RoundParser) MergeResult(other *result.Result) {
	rp.res.Merge(other)
}

// Parse processes a complete round result and updates simulation statistics
func (rp *RoundParser) Parse(roundResult *resp.Round) error {
	var totalFreeWin uint64

	totalBaseWin := rp.parseBasePhase(roundResult.BasePhase)

	if len(roundResult.FreePhase) > 0 {
		totalFreeWin = rp.parseFreePhase(roundResult.FreePhase)

		rp.res.Base().AddFreeGameSteps(len(roundResult.FreePhase))
	}

	roundTotalWin := totalBaseWin + totalFreeWin

	rp.updateRoundStatistics(roundTotalWin)

	return nil
}

// parseBasePhase processes base game phase results
func (rp *RoundParser) parseBasePhase(baseResult *resp.Base) uint64 {
	baseTotalWin := baseResult.TotalWin()

	rp.res.Base().AddWins(baseTotalWin)
	rp.res.BaseWinRanges.AddWinRange(baseTotalWin)

	if baseTotalWin > 0 {
		rp.res.BasePositiveBets().Add()
	}

	if baseResult.MaxWinReached() {
		rp.res.AddMaxWinReached()
	}

	return baseTotalWin
}

// parseFreePhase processes all free spin results
func (rp *RoundParser) parseFreePhase(freeSteps []*resp.Free) uint64 {
	var totalFreeWin uint64

	for _, freeStep := range freeSteps {
		stepWin := rp.processSingleFreeStep(freeStep)
		totalFreeWin += stepWin
	}

	rp.res.FreeWinRanges.AddWinRange(totalFreeWin)

	return totalFreeWin
}

// processSingleFreeStep processes a single free spin step
func (rp *RoundParser) processSingleFreeStep(freeResult *resp.Free) uint64 {
	stepWin := freeResult.StepWin()

	rp.res.Free().AddWins(stepWin)

	if freeResult.MaxWinReached() {
		rp.res.AddMaxWinReached()
	}

	return stepWin
}

// updateRoundStatistics updates overall round statistics
func (rp *RoundParser) updateRoundStatistics(totalWin uint64) {
	if totalWin > 0 {
		rp.res.Xi().Add(totalWin)
		rp.res.AddMaxTotalWin(totalWin)
		rp.res.PositiveBets().Add()
	}
}

// GetResult returns the accumulated simulation results
func (rp *RoundParser) GetResult() *result.Result {
	return rp.res
}
