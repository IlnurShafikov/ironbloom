package datatransfer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// costMultiplier is the cost multiplier of the simple game mode.
const costMultiplier = 20

// testBoard returns a small board that is valid for the transformation.
func testBoard() model.Board {
	return model.Board{
		{rules.Ace, rules.High1, rules.High2},
		{rules.Jack, rules.King, rules.Med1},
		{rules.Ace, rules.Queen, rules.Med2},
	}
}

// freeSpin describes a free spin of a test round.
type freeSpin struct {
	// stepWin is the win of the spin in base units.
	stepWin uint64
	// reTriggers is the number of free spins added by the spin.
	reTriggers uint32
	// maxWinReached marks the spin that hit the max win limit.
	maxWinReached bool
}

// buildRound creates a round the way the game phases do: the total win
// of every phase is cumulative and includes the wins of all previous phases.
func buildRound(baseWin uint64, fsCount uint32, spins []freeSpin) *result.Round {
	round := result.NewRound(costMultiplier)

	base := result.NewBase(testBoard())
	base.SetTotalWin(baseWin)
	base.SetFreeSpinsInitialsCount(fsCount)
	round.BasePhase = base

	total := baseWin

	for _, spin := range spins {
		total += spin.stepWin

		free := result.NewFree(testBoard(), nil)
		free.SetStepWin(spin.stepWin)
		free.SetTotalWin(total)
		free.SetReTriggers(spin.reTriggers)
		free.SetMaxWinReached(spin.maxWinReached)

		round.FreePhase = append(round.FreePhase, free)
	}

	return round
}

// TestTransfer_FreeSpinsTotalWinIsCumulative is a regression test: the total win
// of every free step sent to the client must include the wins of all previous
// free steps, so the last step carries the total win of the whole round.
func TestTransfer_FreeSpinsTotalWinIsCumulative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseWin   uint64
		bet       uint64
		fsCount   uint32
		spins     []freeSpin
		wantTotal []uint64
		wantStep  []uint64
	}{
		{
			name:      "base win is carried through all free spins",
			baseWin:   10,
			bet:       1,
			fsCount:   3,
			spins:     []freeSpin{{stepWin: 5}, {stepWin: 0}, {stepWin: 7}},
			wantTotal: []uint64{10, 15, 15, 22},
			wantStep:  []uint64{10, 5, 0, 7},
		},
		{
			name:      "no base win, win only in the middle of the bonus",
			baseWin:   0,
			bet:       1,
			fsCount:   3,
			spins:     []freeSpin{{stepWin: 0}, {stepWin: 220}, {stepWin: 0}},
			wantTotal: []uint64{0, 0, 220, 220},
			wantStep:  []uint64{0, 0, 220, 0},
		},
		{
			name:      "all values are scaled by the bet",
			baseWin:   4,
			bet:       5,
			fsCount:   2,
			spins:     []freeSpin{{stepWin: 6}, {stepWin: 1}},
			wantTotal: []uint64{20, 50, 55},
			wantStep:  []uint64{20, 30, 5},
		},
		{
			name:      "retriggered spins keep accumulating",
			baseWin:   2,
			bet:       1,
			fsCount:   1,
			spins:     []freeSpin{{stepWin: 3, reTriggers: 2}, {stepWin: 4}, {stepWin: 5}},
			wantTotal: []uint64{2, 5, 9, 14},
			wantStep:  []uint64{2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			round := buildRound(tt.baseWin, tt.fsCount, tt.spins)

			res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, tt.bet)
			require.NoError(t, err)
			require.Len(t, res.Steps, len(tt.wantTotal))

			for i, cliStep := range res.Steps {
				require.Equal(t, tt.wantTotal[i], cliStep.TotalWin, "total win of step %d", i)
				require.Equal(t, tt.wantStep[i], cliStep.StepWin, "step win of step %d", i)
			}

			last := res.Steps[len(res.Steps)-1]
			require.Equal(t, round.FreePhase[len(round.FreePhase)-1].TotalWin()*tt.bet, last.TotalWin,
				"last step must carry the total win of the round")
		})
	}
}

// TestTransfer_FreeSpinsStopAtMaxWin checks that the steps after the one
// that reached the max win are not sent to the client and that the total win
// of the last sent step stays cumulative.
func TestTransfer_FreeSpinsStopAtMaxWin(t *testing.T) {
	t.Parallel()

	round := buildRound(10, 4, []freeSpin{
		{stepWin: 5},
		{stepWin: 100, maxWinReached: true},
		{stepWin: 50},
	})

	res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, 1)
	require.NoError(t, err)
	require.Len(t, res.Steps, 3)

	last := res.Steps[len(res.Steps)-1]
	require.True(t, last.MaxWinReached)
	require.Equal(t, uint64(115), last.TotalWin)
}

// TestTransfer_WithoutFreeSpins checks that a round without a bonus
// consists of the single base step.
func TestTransfer_WithoutFreeSpins(t *testing.T) {
	t.Parallel()

	round := buildRound(12, 0, nil)

	res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, 3)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1)
	require.Equal(t, uint64(36), res.Steps[0].TotalWin)
	require.Equal(t, uint64(36), res.Steps[0].StepWin)
}
