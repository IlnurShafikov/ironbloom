package result

// BaseGame accumulates statistics for the base game phase across a simulation run.
type BaseGame struct {
	// Wins is the sum of all win amounts from base game spins.
	Wins uint64
	// FreeGameTriggered is the number of rounds that triggered the free game.
	FreeGameTriggered uint64
	// AllFreeGameSteps is the total number of free spins awarded by the base game.
	AllFreeGameSteps int
}

// newBaseGame creates a BaseGame with zeroed statistics.
func newBaseGame() *BaseGame {
	return &BaseGame{
		Wins:              0,
		FreeGameTriggered: 0,
		AllFreeGameSteps:  0,
	}
}

// AddWins adds the given win amount to the base game win total.
func (b *BaseGame) AddWins(waysWins uint64) {
	b.Wins += waysWins
}

// AddFreeGameSteps records the free spins awarded by a base spin.
// Spins without free spins (steps <= 0) are ignored.
//
// Parameters:
//   - steps: number of free spins awarded
func (b *BaseGame) AddFreeGameSteps(steps int) {
	if steps > 0 {
		b.AllFreeGameSteps += steps
		b.FreeGameTriggered++
	}
}

// GetWins returns the total win amount accumulated from base game spins.
func (b *BaseGame) GetWins() uint64 {
	return b.Wins
}

// GetFreeGameTriggered returns the number of rounds that triggered the free game.
func (b *BaseGame) GetFreeGameTriggered() uint64 {
	return b.FreeGameTriggered
}

// GetAllFreeGameSteps returns the total number of free spins awarded by the base game.
func (b *BaseGame) GetAllFreeGameSteps() int {
	return b.AllFreeGameSteps
}

// Merge adds another BaseGame's statistics into this one.
func (b *BaseGame) Merge(other *BaseGame) {
	b.Wins += other.Wins
	b.AllFreeGameSteps += other.AllFreeGameSteps
	b.FreeGameTriggered += other.FreeGameTriggered
}
