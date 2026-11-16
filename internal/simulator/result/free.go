package result

type FreeGame struct {
	Wins uint64
}

func newFreeGame() *FreeGame {
	return &FreeGame{
		Wins: 0,
	}
}

func (f *FreeGame) AddWins(wins uint64) {
	f.Wins += wins
}

func (f *FreeGame) GetWins() uint64 {
	return f.Wins
}

func (f *FreeGame) Merge(other *FreeGame) {
	f.Wins += other.Wins
}
