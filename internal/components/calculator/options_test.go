package calculator

import (
	"testing"

	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestWithWaysWinMultiplier(t *testing.T) {
	newPos := func(reelIndex int, rowIndex int) model.Position {
		return model.Position{ReelIndex: reelIndex, RowIndex: rowIndex}
	}

	tests := []struct {
		name       string
		wilds      []result.WildMultiplier
		win        calculators.WaysWin
		wantPayout uint64
	}{
		{
			name:  "without wilds",
			wilds: []result.WildMultiplier{},
			win: calculators.WaysWin{
				Positions: model.Positions{
					newPos(0, 0),
					newPos(1, 0),
					newPos(2, 0),
				},
				PayoutPerWay: 10,
				TotalPayout:  10,
			},
			wantPayout: 10,
		},
		{
			name: "with wilds but not on winning reels",
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(newPos(3, 0), 2, 0),
			},
			win: calculators.WaysWin{
				Positions: model.Positions{
					newPos(0, 1),
					newPos(1, 0),
					newPos(2, 0),
				},
				PayoutPerWay: 5,
				TotalPayout:  5,
			},
			wantPayout: 5,
		},
		{
			name: "with one wild on winning reel",
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(
					newPos(1, 0),
					3, 0),
			},
			win: calculators.WaysWin{
				Positions: model.Positions{
					newPos(0, 0),
					newPos(1, 0),
					newPos(2, 0),
				},
				PayoutPerWay: 10,
				TotalPayout:  10,
				Ways:         1,
			},
			wantPayout: 30,
		},
		{
			name: "with 3 wilds",
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(newPos(0, 0), 2, 0),
				result.NewWildMultiplier(newPos(0, 1), 3, 0),
				result.NewWildMultiplier(newPos(2, 0), 5, 0),
			},
			// 1ways: x2 + 0 + x5
			// 2ways: x3 + 0 + x5
			// [w, w, s]
			// [1, x, x]
			// [w, x, x]
			win: calculators.WaysWin{
				Positions: model.Positions{
					newPos(0, 0),
					newPos(0, 1),
					newPos(1, 0),
					newPos(2, 0),
				},
				PayoutPerWay: 10,
				TotalPayout:  10 * 2,
				Ways:         2,
			},
			wantPayout: 250, // (10*2*1*5) + (10*3*1*5) = 100 + 150
		},
		{
			name: "multiple wilds on different reels",
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(newPos(0, 0), 2, 0),
				result.NewWildMultiplier(newPos(2, 0), 4, 0), // ignoring this one
			},
			win: calculators.WaysWin{
				Positions: model.Positions{
					newPos(0, 1),
					newPos(1, 0),
				},
				PayoutPerWay: 5,
				TotalPayout:  5,
				Ways:         1,
			},
			wantPayout: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := WithWaysWinMultiplier(tt.wilds)
			gotWin, gotStop := strategy(tt.win)

			require.Equal(t, tt.wantPayout, gotWin.TotalPayout)
			require.False(t, gotStop)
		})
	}
}

func TestCountSymbolOnReel(t *testing.T) {
	tests := []struct {
		name      string
		reel      int
		positions model.Positions
		wantCount uint64
	}{
		{
			name:      "no symbols on reel",
			reel:      0,
			positions: model.Positions{},
			wantCount: 0,
		},
		{
			name: "one symbol on reel",
			reel: 0,
			positions: model.Positions{
				{ReelIndex: 0, RowIndex: 0},
			},
			wantCount: 1,
		},
		{
			name: "multiple symbols on reel",
			reel: 1,
			positions: model.Positions{
				{ReelIndex: 1, RowIndex: 0},
				{ReelIndex: 1, RowIndex: 1},
				{ReelIndex: 1, RowIndex: 2},
			},
			wantCount: 3,
		},
		{
			name: "symbols on different reels",
			reel: 2,
			positions: model.Positions{
				{ReelIndex: 0, RowIndex: 0},
				{ReelIndex: 1, RowIndex: 0},
				{ReelIndex: 2, RowIndex: 0},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSymbolOnReel(tt.reel, tt.positions)
			if got != tt.wantCount {
				t.Errorf("countSymbolOnReel() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestCountWildsOnReel(t *testing.T) {
	tests := []struct {
		name      string
		reel      int
		wilds     []result.WildMultiplier
		wantCount uint64
	}{
		{
			name:      "no wilds",
			reel:      0,
			wilds:     []result.WildMultiplier{},
			wantCount: 0,
		},
		{
			name: "wild with multiplier 1",
			reel: 0,
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(model.Position{ReelIndex: 0, RowIndex: 0}, 1, 0),
			},
			wantCount: 0, // multiplier - 1 = 0
		},
		{
			name: "wild with multiplier 3",
			reel: 0,
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(model.Position{ReelIndex: 0, RowIndex: 0}, 3, 0),
			},
			wantCount: 2, // multiplier - 1 = 2
		},
		{
			name: "multiple wilds on same reel",
			reel: 1,
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(model.Position{ReelIndex: 1, RowIndex: 0}, 2, 0),
				result.NewWildMultiplier(model.Position{ReelIndex: 1, RowIndex: 1}, 4, 0),
			},
			wantCount: 4, // (2-1) + (4-1) = 1 + 3 = 4
		},
		{
			name: "wilds on different reels",
			reel: 2,
			wilds: []result.WildMultiplier{
				result.NewWildMultiplier(model.Position{ReelIndex: 0, RowIndex: 0}, 3, 0),
				result.NewWildMultiplier(model.Position{ReelIndex: 1, RowIndex: 0}, 2, 0),
				result.NewWildMultiplier(model.Position{ReelIndex: 2, RowIndex: 0}, 5, 0),
			},
			wantCount: 4, // only wild on reel 2: 5-1 = 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countWildsOnReel(tt.reel, tt.wilds)
			if got != tt.wantCount {
				t.Errorf("countWildsOnReel() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}
