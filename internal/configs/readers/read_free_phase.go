package readers

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/wild"
	"github.com/IlnurShafikov/ironbloom/internal/configs"
	"github.com/IlnurShafikov/ironbloom/internal/configs/readers/common"
	"github.com/IlnurShafikov/ironbloom/internal/game/phase/free"
	"github.com/IlnurShafikov/ironbloom/internal/tools"
	"github.com/IlnurShafikov/slot-engine/pkg/generator"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/reader"
)

// readFreePhase reads and constructs the free spins phase configuration.
// Currently returns a basic configuration; extend as needed for your game.
//
// Parameters:
//   - csvReader: CSV file adapter for reading configuration files
//   - modPath: path builder pointing to the game mode directory
//   - general: general configuration shared across all phases
//
// Returns:
//   - pointer to SpinPhase containing free spins configuration
//   - error if reading or parsing fails
func readFreePhase(
	csvReader *reader.Adapter,
	modPath *tools.PathBuilder,
	general *configs.GeneralConfig,
) (*free.SpinPhase, error) {
	const (
		freePrefix = "free"
	)

	phasePath := modPath.Phase(freePrefix)

	dist, err := common.ReadDistributions(csvReader, phasePath)
	if err != nil {
		return nil, fmt.Errorf("read distributions: %w", err)
	}

	reelsGenerator, err := readFreeReelsGenerator(csvReader, phasePath, general, dist.CopyValues())
	if err != nil {
		return nil, fmt.Errorf("read reels generator: %w", err)
	}

	calc := readWaysCalculator(general)
	multiGetter := wild.NewMultiplierGetter(general.WildIDs, general.Multipliers.WildMultipliers)

	spinPhase := free.NewSpinPhase(
		reelsGenerator,
		dist,
		multiGetter,
		calc,
		general.FsInitCounter,
	)

	return spinPhase, nil
}

// readFreeReelsGenerator reads reel-strip data for every trigger group and
// builds the board generator used for free spins.
//
// Parameters:
//   - csvReader: CSV file adapter backed by the embedded config FS
//   - phasePath: path builder pointing to the free phase directory
//   - general: shared game configuration supplying the per-reel symbol counts
//   - groupIDs: group IDs for which reel files must be loaded
//
// Returns:
//   - pointer to Generator with all group reel strips loaded
//   - error if any reel file cannot be read or the generator cannot be initialised
func readFreeReelsGenerator(
	csvReader *reader.Adapter,
	phasePath *tools.PathBuilder,
	general *configs.GeneralConfig,
	groupIDs []model.GroupID,
) (*generator.Generator, error) {
	reels, err := common.ReadReels(csvReader, phasePath, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("read reels: %w", err)
	}

	reelGen, err := generator.NewReelsGenerator(reels, general.FreeSymbolsOnReels)
	if err != nil {
		return nil, fmt.Errorf("make base board generator: %w", err)
	}

	return reelGen, nil
}
