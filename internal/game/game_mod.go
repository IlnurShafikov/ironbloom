package game

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/limiter"
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// BasePhase defines the interface for base game phase operations
type BasePhase interface {
	// Spin executes a single base game spin.
	Spin(rng model.RNG, winLimiter limiter.WinLimiter) (*result.Base, error)
}

// FreePhase defines the interface for free spins phase operations
type FreePhase interface {
	// Spins executes all free spins, including the ones added by retriggers.
	Spins(
		rng model.RNG,
		fsCount uint32,
		winLimiter limiter.WinLimiter,
	) ([]*result.Free, error)
}

// Mod represents a complete game modification with base and free spin phases
type Mod struct {
	// baseSpinPhase handles base game spin logic
	baseSpinPhase BasePhase

	// freeSpinPhase handles free spins logic
	freeSpinPhase FreePhase

	// costMultiplier is the multiplier applied to bet values
	costMultiplier uint64

	// maxWinLimit is the maximum win amount allowed per round
	maxWinLimit uint64
}

// NewMod creates a new Mod instance with the specified phases and limits.
//
// Parameters:
//   - basePhase: base game phase implementation
//   - freeSpinPhase: free spins phase implementation
//   - costMultiplier: multiplier for bet values
//   - maxWinLimit: maximum win amount allowed per round
//
// Returns:
//   - pointer to the newly created Mod
func NewMod(
	basePhase BasePhase,
	freeSpinPhase FreePhase,
	costMultiplier, maxWinLimit uint64,
) *Mod {
	return &Mod{
		baseSpinPhase:  basePhase,
		freeSpinPhase:  freeSpinPhase,
		costMultiplier: costMultiplier,
		maxWinLimit:    maxWinLimit,
	}
}

// CostMultiplier returns the cost multiplier for this game mod
func (g *Mod) CostMultiplier() uint64 {
	return g.costMultiplier
}

// Spin executes a complete game round and returns the result
func (g *Mod) Spin(rng model.RNG) (*result.Round, error) {
	winLimiter, err := limiter.New(g.maxWinLimit)
	if err != nil {
		return nil, fmt.Errorf("initialize maxWinCollector %w", err)
	}

	roundResult := result.NewRound(g.costMultiplier)

	if err := g.startBasePhase(rng, roundResult, winLimiter); err != nil {
		return nil, fmt.Errorf("execute base phase: %w", err)
	}

	if err := g.startFreePhaseIfTriggered(rng, roundResult, winLimiter); err != nil {
		return nil, fmt.Errorf("execute free phase: %w", err)
	}

	return roundResult, nil
}

// startBasePhase executes the base game phase and updates the round result
func (g *Mod) startBasePhase(
	rng model.RNG, roundResult *result.Round, winLimiter limiter.WinLimiter,
) error {
	baseResult, err := g.baseSpinPhase.Spin(rng, winLimiter)
	if err != nil {
		return fmt.Errorf("spin base game: %w", err)
	}

	roundResult.BasePhase = baseResult

	return nil
}

// startFreePhaseIfTriggered checks for free spin triggers and executes the free spin phase
func (g *Mod) startFreePhaseIfTriggered(
	rng model.RNG, roundResult *result.Round, winLimiter limiter.WinLimiter,
) error {
	fsCount := roundResult.BasePhase.FreeSpinsInitialsCount()
	if fsCount == 0 {
		return nil // No free spins triggered
	}

	freeResults, err := g.freeSpinPhase.Spins(rng, fsCount, winLimiter)
	if err != nil {
		return fmt.Errorf("spin free games (count: %d): %w", fsCount, err)
	}

	roundResult.FreePhase = freeResults

	return nil
}
