package configs

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/pairs"
	"github.com/IlnurShafikov/slot-engine/pkg/paytable"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/fscounter"
)

// GeneralConfig holds the general game configuration that is shared across all game modes.
// It includes payout tables and various multipliers used throughout the game.
type GeneralConfig struct {
	// PayTable contains symbol payouts and payout ranges for the game
	PayTable *paytable.PayTable

	// Multipliers contains various multiplier configurations (cost, max win, etc.)
	Multipliers *MultipliersConfigs

	// BaseSymbolsOnReels is the number of visible symbol rows per reel in base phase.
	BaseSymbolsOnReels []int

	// FreeSymbolsOnReels is the number of visible symbol rows per reel in free phase.
	FreeSymbolsOnReels []int

	WildIDs []int

	FsInitCounter fscounter.FSInitCounter

	// Volatility represents the game's volatility setting, affecting risk and payout frequency.
	Volatility string
}

// MultipliersConfigs contains all multiplier-related configuration values.
type MultipliersConfigs struct {
	WildMultipliers pairs.Pairs

	// MaxWinMultiplier defines the maximum win as a multiplier of the bet
	MaxWinMultiplier uint64

	// SimpleCostMultiplier defines the cost multiplier for the simple game mode
	SimpleCostMultiplier uint64
}

// MaxWinLimit calculates the absolute maximum win limit for the game.
// The limit is calculated as MaxWinMultiplier * SimpleCostMultiplier.
//
// Returns:
//   - the maximum win limit value
func (c *GeneralConfig) MaxWinLimit() uint64 {
	return c.Multipliers.MaxWinMultiplier * c.SimpleCostMultiplier()
}

// SimpleCostMultiplier returns the cost multiplier for the simple game mode.
//
// Returns:
//   - the simple mode cost multiplier
func (c *GeneralConfig) SimpleCostMultiplier() uint64 {
	return c.Multipliers.SimpleCostMultiplier
}
