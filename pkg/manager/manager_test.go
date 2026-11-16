package manager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/step"
	"github.com/IlnurShafikov/ironbloom/pkg/manager"
)

const (
	rtp96 = 96
	bet   = 2
)

// TestGameManager_GetRoundResult tests the GetRoundResult method of the GameManager.
func TestGameManager_GetRoundResult(t *testing.T) {
	gameManager, err := manager.NewGameManager()
	require.NoError(t, err)

	for range 1000 {
		spinRes, err := gameManager.GetRoundResult(
			context.Background(), rules.SimpleModID, rules.EmptyFeatureID, rtp96, seed(), bet)
		require.NoError(t, err)
		require.NotNil(t, spinRes)
	}
}

// TestGameManager_AllRtpGetRoundResult tests the GetRoundResult method for all rtp versions
// of the GameManager.
func TestGameManager_AllRtpGetRoundResult(t *testing.T) {
	rtps := []int{rtp96}

	gameManager, err := manager.NewGameManager()
	require.NoError(t, err)

	for _, curRtp := range rtps {
		for range 100 {
			spinRes, err := gameManager.GetRoundResult(
				context.Background(), rules.SimpleModID, rules.EmptyFeatureID, curRtp, seed(), bet)
			require.NoError(t, err)
			require.NotNil(t, spinRes)
		}
	}
}

// TestEnterGame tests the EnterGame method of the GameManager.
// It ensures that the method can be called without errors and returns a valid response.
func TestEnterGame(t *testing.T) {
	gm, err := manager.NewGameManager()
	require.NoError(t, err)

	ctx := context.Background()
	enterGame, err := gm.EnterGame(ctx, bet, rtp96)
	require.NoError(t, err)
	require.NotNil(t, enterGame)
}

// TestAllGameMods tests all game to ensure they can be processed without errors.
func TestAllGameMods(t *testing.T) {
	gameManager, err := manager.NewGameManager()
	require.NoError(t, err)

	allGameMods := []string{
		rules.SimpleModID,
	}

	for _, modID := range allGameMods {
		for range 1000 {
			spinRes, err := gameManager.GetRoundResult(
				context.Background(), modID, rules.EmptyFeatureID, rtp96, seed(), bet)
			require.NoError(t, err)
			require.NotNil(t, spinRes)
		}
	}
}

// TestGameManager_GetRoundResult_Payout checks that the total win sent to the client
// equals the total win of the internal round result scaled by the bet.
func TestGameManager_GetRoundResult_Payout(t *testing.T) {
	gameManager, err := manager.NewGameManager()
	require.NoError(t, err)

	for range 1000 {
		gameSeed := seed()

		beforeDT, err := gameManager.GenerateRoundResult(rtp96, rules.SimpleModID, rules.EmptyFeatureID, gameSeed)
		require.NoError(t, err)
		require.NotNil(t, beforeDT)

		afterDT, err := gameManager.GetRoundResult(
			context.Background(), rules.SimpleModID, rules.EmptyFeatureID, rtp96, gameSeed, bet)
		require.NoError(t, err)
		require.NotNil(t, afterDT)

		rawTotalWin := calculateRawTotalWin(beforeDT)
		transferTotalWin := calculateTransferTotalWin(afterDT.Steps)

		resMsg := fmt.Sprintf(
			"mismatch win %d need %d, seed=%d",
			transferTotalWin, rawTotalWin*bet, afterDT.Seed,
		)

		require.Equal(t, rawTotalWin*bet, transferTotalWin, resMsg)
	}
}

// calculateRawTotalWin calculates the total win from the internal game result
// before it's been transformed or scaled by bet value.
// The total win is cumulative, so it is taken from the last executed phase.
//
// Parameters:
//   - roundRes: the internal round result from the game engine
//
// Returns:
//   - total win amount (in base units, not scaled by bet)
func calculateRawTotalWin(roundRes *result.Round) uint64 {
	if freeSpins := roundRes.FreePhase; len(freeSpins) > 0 {
		return freeSpins[len(freeSpins)-1].TotalWin()
	}

	return roundRes.BasePhase.TotalWin()
}

// calculateTransferTotalWin calculates the total win from the transformed
// client-facing steps. This value should be scaled by the bet value.
// The total win of a step is cumulative, so it is taken from the last step.
//
// Parameters:
//   - allSteps: slice of transformed steps sent to the client
//
// Returns:
//   - total win amount (already scaled by bet value)
func calculateTransferTotalWin(allSteps []*step.Step) uint64 {
	if len(allSteps) == 0 {
		return 0
	}

	return allSteps[len(allSteps)-1].TotalWin
}

// seed generates a non-zero seed based on the current time.
// This ensures each test run uses different random seeds.
//
// Returns:
//   - a non-zero int64 seed value for RNG
func seed() int64 {
	var respSeed int64

	for {
		respSeed = int64(time.Now().Nanosecond())

		if respSeed != 0 {
			break
		}
	}

	return respSeed
}

// TestGameManager_GetRoundResult_Deterministic checks that the same seed
// always produces the same sequence of steps for every RTP version.
func TestGameManager_GetRoundResult_Deterministic(t *testing.T) {
	gameManager, err := manager.NewGameManager()
	require.NoError(t, err)

	for _, rtp := range []int{90, 92, 94, 96} {
		for gameSeed := int64(1); gameSeed <= 300; gameSeed++ {
			first, err := gameManager.GetRoundResult(
				context.Background(), rules.SimpleModID, rules.EmptyFeatureID, rtp, gameSeed, bet)
			require.NoError(t, err)

			second, err := gameManager.GetRoundResult(
				context.Background(), rules.SimpleModID, rules.EmptyFeatureID, rtp, gameSeed, bet)
			require.NoError(t, err)

			require.Len(t, second.Steps, len(first.Steps), "rtp=%d seed=%d", rtp, gameSeed)

			for i := range first.Steps {
				require.Equal(t, first.Steps[i].StepWin, second.Steps[i].StepWin,
					"rtp=%d seed=%d step=%d", rtp, gameSeed, i)
				require.Equal(t, first.Steps[i].TotalWin, second.Steps[i].TotalWin,
					"rtp=%d seed=%d step=%d", rtp, gameSeed, i)
			}
		}
	}
}
