package result

// FreeGame accumulates statistics for the free game phase across a simulation run.
type FreeGame struct {
	// Wins is the sum of all win amounts from free game spins.
	Wins uint64
}

// newFreeGame creates a FreeGame with zeroed statistics.
func newFreeGame() *FreeGame {
	return &FreeGame{
		Wins: 0,
	}
}

// AddWins adds the given win amount to the free game win total.
func (f *FreeGame) AddWins(wins uint64) {
	f.Wins += wins
}

// GetWins returns the total win amount accumulated from free game spins.
func (f *FreeGame) GetWins() uint64 {
	return f.Wins
}

// Merge adds another FreeGame's statistics into this one.
func (f *FreeGame) Merge(other *FreeGame) {
	f.Wins += other.Wins
}
