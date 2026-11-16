package result

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// Free represents the result of a single free spin.
// It contains all information about the free spin outcome, including events, wins, and retriggers.
type Free struct {
	Board model.Board
	// events accumulates all events that occurred during this free spin.
	events      []events.Event
	Multipliers []WildMultiplier
	// totalWin is the cumulative win amount up to and including this free spin.
	totalWin uint64
	// stepWin is the win amount for this specific free spin.
	stepWin uint64
	// reTriggers is the number of additional free spins triggered by this spin.
	reTriggers uint32
	// maxWinReached indicates if the maximum win limit was reached.
	maxWinReached bool
}

// NewFree creates a new Free spin result initialized with the provided board and expansion state.
//
// Returns:
//   - pointer to a new Free instance with counters and flags initialized to zero/false
func NewFree(board model.Board, multipliers []WildMultiplier) *Free {
	return &Free{
		Board:         board,
		Multipliers:   multipliers,
		events:        nil,
		totalWin:      0,
		stepWin:       0,
		reTriggers:    0,
		maxWinReached: false,
	}
}

// SetEvents sets the events that occurred during this free spin.
//
// Parameters:
//   - e: slice of events to set
func (f *Free) SetEvents(e []events.Event) {
	f.events = e
}

// SetReTriggers sets the number of additional free spins triggered.
//
// Parameters:
//   - count: number of additional free spins
func (f *Free) SetReTriggers(count uint32) {
	f.reTriggers = count
}

// SetTotalWin sets the cumulative win amount.
//
// Parameters:
//   - win: cumulative win amount
func (f *Free) SetTotalWin(win uint64) {
	f.totalWin = win
}

// SetMaxWinReached sets whether the maximum win limit was reached.
//
// Parameters:
//   - reached: true if max win was reached, false otherwise
func (f *Free) SetMaxWinReached(reached bool) {
	f.maxWinReached = reached
}

// SetStepWin sets the win amount for this specific free spin.
//
// Parameters:
//   - stepWin: win amount for this spin
func (f *Free) SetStepWin(stepWin uint64) {
	f.stepWin = stepWin
}

// Events returns all events that occurred during this free spin.
//
// Returns:
//   - slice of events
func (f *Free) Events() []events.Event {
	return f.events
}

// ReTriggers returns the number of additional free spins triggered.
//
// Returns:
//   - number of retriggers
func (f *Free) ReTriggers() uint32 {
	return f.reTriggers
}

// TotalWin returns the cumulative win amount.
//
// Returns:
//   - cumulative win amount
func (f *Free) TotalWin() uint64 {
	return f.totalWin
}

// MaxWinReached returns whether the maximum win limit was reached.
//
// Returns:
//   - true if max win was reached, false otherwise
func (f *Free) MaxWinReached() bool {
	return f.maxWinReached
}

// StepWin returns the win amount for this specific free spin.
//
// Returns:
//   - win amount for this spin
func (f *Free) StepWin() uint64 {
	return f.stepWin
}
