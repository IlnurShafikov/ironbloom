package datatransfer

import (
	"fmt"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"

	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/common"
	freeDto "github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/free"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/step"
)

// TransformFreeStep transforms all free spins from a round into client-facing steps.
// It processes each free spin sequentially, tracking counters and respins.
//
// Parameters:
//   - res: the complete round result containing free spins
//   - allCliSteps: accumulated list of client steps
//   - baseParent: base game state to return to when max win is reached
//   - totalRoundWin: cumulative win before free spins
//   - betValue: the player's bet amount
//
// Returns:
//   - updated slice of client steps
//   - total win amount (including free spins)
//   - error if transformation fails
func (t *Transfer) TransformFreeStep(
	res *result.Round, allCliSteps []*step.Step,
	totalRoundWin, betValue uint64,
) ([]*step.Step, error) {
	initialsCount := res.BasePhase.FreeSpinsInitialsCount()

	if initialsCount == 0 {
		return allCliSteps, nil
	}

	freeSteps := res.FreePhase
	counter := initCounter(initialsCount)

	for stepNum := range freeSteps {
		freeRes := freeSteps[stepNum]
		isLastStep := stepNum == len(freeSteps)-1

		cliStep, err := t.processFreeStep(
			freeRes, isLastStep, &counter, betValue, totalRoundWin)
		if err != nil {
			return nil, err
		}

		allCliSteps = append(allCliSteps, cliStep)

		// TotalWin of a free step is cumulative, so the next step builds on this one.
		totalRoundWin = cliStep.TotalWin

		finalizeCounter(&counter)

		if allCliSteps[len(allCliSteps)-1].MaxWinReached {
			break
		}
	}

	return allCliSteps, nil
}

// processFreeStep transforms a single free spin into a client step.
// Handles symbol transformation, event processing, counter updates, and respins.
//
// Parameters:
//   - freeRes: the free spin result to transform
//   - isLastStep: whether this is the final free spin in the sequence
//   - counter: the free spins counter to update
//   - betValue: the player's bet amount
//   - totalWin: cumulative win before this spin
//
// Returns:
//   - pointer to protobuf Step message
//   - win amount for this free spin
//   - error if transformation fails
func (t *Transfer) processFreeStep(
	freeRes *result.Free, isLastStep bool,
	counter *common.Counter, betValue, totalWin uint64,
) (*step.Step, error) {
	stepWin := freeRes.StepWin() * betValue

	symbols, err := TransformSymbols(freeRes.Board, freeRes.Multipliers)
	if err != nil {
		return nil, fmt.Errorf("transfer beforeSymbols: %w", err)
	}

	eventList, err := ExtractAndTransformEvents(freeRes, betValue)
	if err != nil {
		return nil, fmt.Errorf("transform events: %w", err)
	}

	updateCounter(counter, freeRes)

	payload := buildFreeSpinPayload(
		*counter, eventList, symbols, isLastStep)

	currStep := buildFreeSpinStep(payload, totalWin+stepWin, stepWin, freeRes)

	data, err := jsoniter.Marshal(currStep)
	if err != nil {
		return nil, fmt.Errorf("marshal base step: %w", err)
	}

	id, err := uuid.New().MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal step id: %w", err)
	}

	cliStep := makeFreeClientStep(id, data, stepWin, totalWin+stepWin, freeRes)

	return cliStep, nil
}

// initCounter creates a new counter initialized for the start of free spins or respins.
//
// Parameters:
//   - initialsCount: number of spins initially granted
//
// Returns:
//   - Counter with left=initialsCount-1, added=0, played=1
func initCounter(initialsCount uint32) common.Counter {
	return common.Counter{
		Left:   initialsCount - 1,
		Added:  0,
		Played: 1,
	}
}

// updateCounter updates the free spins counter based on the step result.
// Handles retriggers and max win scenarios.
//
// Parameters:
//   - counter: the counter to update
//   - freeRes: the free spin result
func updateCounter(counter *common.Counter, freeRes *result.Free) {
	if freeRes.ReTriggers() > 0 {
		counter.Added = freeRes.ReTriggers()
		counter.Left += freeRes.ReTriggers()
	}

	if freeRes.MaxWinReached() {
		counter.Left = 0
		counter.Added = 0
	}
}

// buildFreeSpinPayload constructs the payload for a free spin step DTO.
//
// Parameters:
//   - counter: current state of the free spins counter
//   - eventList: game events that occurred
//   - symbols: initial symbol board
//   - isLastStep: whether this is the final free spin in the sequence
//
// Returns:
//   - pointer to free Payload DTO
func buildFreeSpinPayload(
	counter common.Counter, eventList []events.Event,
	symbols []*common.Symbol, isLastStep bool,
) *freeDto.Payload {
	return &freeDto.Payload{
		Counter:      counter,
		Events:       eventList,
		Symbols:      symbols,
		NextStepType: nextStepTypeFromFree(isLastStep),
	}
}

// buildFreeSpinStep constructs a complete free spin Step DTO.
//
// Parameters:
//   - payload: the step payload with game data
//   - totalWin: cumulative win amount
//   - stepWin: win amount for this free spin
//   - freeRes: the free spin result
//
// Returns:
//   - pointer to free Step DTO
func buildFreeSpinStep(
	payload *freeDto.Payload, totalWin, stepWin uint64, freeRes *result.Free,
) *freeDto.Step {
	return &freeDto.Step{
		Payload:       *payload,
		TotalWin:      totalWin,
		StepWin:       stepWin,
		Type:          rules.Step_TYPE_FREE_SPIN,
		MaxWinReached: freeRes.MaxWinReached(),
	}
}

// makeFreeClientStep constructs the protobuf Step message for a free spin.
//
// Parameters:
//   - id: unique identifier for the step
//   - data: JSON-serialized step data
//   - stepWin: win amount for this free spin
//   - totalWin: cumulative win amount
//   - freeRes: the free spin result
//
// Returns:
//   - pointer to protobuf Step message
func makeFreeClientStep(
	id, data []byte, stepWin, totalWin uint64, stepFree *result.Free,
) *step.Step {
	return &step.Step{
		Id:             id,
		ClientResponse: data,
		Cost:           0,
		StepWin:        stepWin,
		TotalWin:       totalWin,
		Type:           rules.Step_TYPE_FREE_SPIN,
		StepsAdded:     stepFree.ReTriggers(),
		MaxWinReached:  stepFree.MaxWinReached(),
	}
}

// finalizeCounter updates the counter after processing a free spin.
// Increments played count, decrements remaining, and resets added.
//
// Parameters:
//   - counter: the counter to finalize
func finalizeCounter(counter *common.Counter) {
	counter.Played++
	counter.Left--
	counter.Added = 0
}

// nextStepTypeFromFree determines what type of step comes after this free spin.
//
// Parameters:
//   - isLastStep: whether this is the final free spin in the sequence
//
// Returns:
//   - TYPE_BASE_SPIN if this is the last free spin
//   - TYPE_FREE_SPIN otherwise
func nextStepTypeFromFree(isLastStep bool) int {
	if isLastStep {
		return rules.Step_TYPE_BASE_SPIN
	}

	return rules.Step_TYPE_FREE_SPIN
}
