package datatransfer_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/common"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/step"
)

// clientPayload is the part of the client response that the tests check.
type clientPayload struct {
	// Counter is the counter of the triggered free spins.
	Counter *common.Counter `json:"nextCounter"`
	// FreeCounter is the counter of a free spin step.
	FreeCounter *common.Counter `json:"counter"`
	// Events are the events of the step.
	Events []struct {
		Type string `json:"type"`
		Ways []struct {
			Payout uint64 `json:"payout"`
		} `json:"ways"`
	} `json:"events"`
	// Symbols is the board shown to the player.
	Symbols []json.RawMessage `json:"symbols"`
	// NextStepType is the type of the step that follows.
	NextStepType int `json:"nextStepType"`
}

// clientResponse is the JSON body of a step sent to the client.
type clientResponse struct {
	// Payload contains the game data of the step.
	Payload clientPayload `json:"payload"`
}

// decodeResponse decodes the client response of a step.
func decodeResponse(t *testing.T, cliStep *step.Step) clientResponse {
	t.Helper()

	var resp clientResponse

	require.NoError(t, json.Unmarshal(cliStep.ClientResponse, &resp))

	return resp
}

// TestTransfer_BaseStep checks the fields of the base step for the rounds
// with and without a bonus.
func TestTransfer_BaseStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		baseWin      uint64
		fsCount      uint32
		bet          uint64
		wantTotal    uint64
		wantStepsAdd uint32
		wantNextType int
	}{
		{
			name:         "no win and no bonus",
			baseWin:      0,
			fsCount:      0,
			bet:          1,
			wantTotal:    0,
			wantStepsAdd: 0,
			wantNextType: rules.Step_TYPE_BASE_SPIN,
		},
		{
			name:         "win is scaled by the bet",
			baseWin:      7,
			fsCount:      0,
			bet:          4,
			wantTotal:    28,
			wantStepsAdd: 0,
			wantNextType: rules.Step_TYPE_BASE_SPIN,
		},
		{
			name:         "scatters trigger free spins",
			baseWin:      3,
			fsCount:      8,
			bet:          2,
			wantTotal:    6,
			wantStepsAdd: 8,
			wantNextType: rules.Step_TYPE_FREE_SPIN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			round := buildRound(tt.baseWin, tt.fsCount, nil)

			res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, tt.bet)
			require.NoError(t, err)
			require.Equal(t, tt.bet, res.UserBet)
			require.Equal(t, uint64(costMultiplier), res.CostMultiplier)
			require.NotEmpty(t, res.Steps)

			base := res.Steps[0]
			require.Equal(t, rules.Step_TYPE_BASE_SPIN, base.Type)
			require.Equal(t, uint64(costMultiplier)*tt.bet, base.Cost)
			require.Equal(t, tt.wantTotal, base.TotalWin)
			require.Equal(t, tt.wantTotal, base.StepWin)
			require.Equal(t, tt.wantStepsAdd, base.StepsAdded)
			require.False(t, base.MaxWinReached)

			payload := decodeResponse(t, base).Payload
			require.Equal(t, tt.wantNextType, payload.NextStepType)
			require.Len(t, payload.Symbols, 9)

			if tt.fsCount == 0 {
				require.Nil(t, payload.Counter)

				return
			}

			require.Equal(t, &common.Counter{Added: tt.fsCount, Left: 0, Played: 0}, payload.Counter)
		})
	}
}

// TestTransfer_BaseStepMaxWin checks that the max win flag reaches the client.
func TestTransfer_BaseStepMaxWin(t *testing.T) {
	t.Parallel()

	round := buildRound(200000, 0, nil)
	round.BasePhase.SetMaxWinReached(true)

	res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, 1)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1)
	require.True(t, res.Steps[0].MaxWinReached)
}

// TestTransfer_EventsAreScaledByBet checks that the payouts of the wins events
// are multiplied by the bet in both base and free steps.
func TestTransfer_EventsAreScaledByBet(t *testing.T) {
	t.Parallel()

	const bet = 3

	round := buildRound(10, 1, []freeSpin{{stepWin: 5}})

	round.BasePhase.SetEvents([]events.Event{newWinsEvent(10)})
	round.FreePhase[0].SetEvents([]events.Event{newWinsEvent(5)})

	res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, bet)
	require.NoError(t, err)
	require.Len(t, res.Steps, 2)

	wantPayouts := []uint64{10 * bet, 5 * bet}

	for i, cliStep := range res.Steps {
		payload := decodeResponse(t, cliStep).Payload
		require.Len(t, payload.Events, 1, "step %d", i)
		require.Equal(t, string(events.WinsType), payload.Events[0].Type)
		require.Len(t, payload.Events[0].Ways, 1)
		require.Equal(t, wantPayouts[i], payload.Events[0].Ways[0].Payout, "step %d", i)
	}
}

// TestTransfer_FreeStepCounters checks the types, counters and the bonus
// related fields of the free steps, including a retrigger.
func TestTransfer_FreeStepCounters(t *testing.T) {
	t.Parallel()

	round := buildRound(0, 2, []freeSpin{
		{stepWin: 1, reTriggers: 8},
		{stepWin: 1},
	})

	res, err := datatransfer.NewTransfer(costMultiplier).Transfer(round, 1)
	require.NoError(t, err)
	require.Len(t, res.Steps, 3)

	first, second := res.Steps[1], res.Steps[2]

	require.Equal(t, rules.Step_TYPE_FREE_SPIN, first.Type)
	require.Equal(t, rules.Step_TYPE_FREE_SPIN, second.Type)
	require.Zero(t, first.Cost, "free spins are free")
	require.Equal(t, uint32(8), first.StepsAdded)
	require.Zero(t, second.StepsAdded)

	firstPayload := decodeResponse(t, first).Payload
	require.Equal(t, &common.Counter{Added: 8, Left: 9, Played: 1}, firstPayload.FreeCounter)
	require.Equal(t, rules.Step_TYPE_FREE_SPIN, firstPayload.NextStepType)

	secondPayload := decodeResponse(t, second).Payload
	require.Equal(t, &common.Counter{Added: 0, Left: 8, Played: 2}, secondPayload.FreeCounter)
	require.Equal(t, rules.Step_TYPE_BASE_SPIN, secondPayload.NextStepType,
		"the last step hands the control back to the base game")
}

// newWinsEvent creates a wins event with a single ways win.
func newWinsEvent(payout uint64) *events.WinsEvent {
	return &events.WinsEvent{
		Event:     events.WinsType,
		Positions: nil,
		WaysWins: []events.WaysWin{
			{PosIndices: nil, Payout: payout, Ways: 1, Symbol: rules.Ace},
		},
	}
}
