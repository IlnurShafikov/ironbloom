package readers

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/components/calculator"
	"github.com/IlnurShafikov/ironbloom/internal/configs"
	"github.com/IlnurShafikov/ironbloom/internal/configs/readers/common"
	"github.com/IlnurShafikov/ironbloom/internal/entities/rules"
	"github.com/IlnurShafikov/ironbloom/internal/game/phase/base"
	"github.com/IlnurShafikov/ironbloom/internal/tools"
	"github.com/IlnurShafikov/slot-engine/pkg/calculators"
	"github.com/IlnurShafikov/slot-engine/pkg/generator"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/reader"
)

// readBasePhase reads and constructs the base game phase configuration.
// It loads symbol distributions and weight groups from CSV files.
//
// Parameters:
//   - csvReader: CSV file adapter for reading configuration files
//   - modPath: path builder pointing to the game mode directory
//   - general: general configuration shared across all phases
//
// Returns:
//   - pointer to SpinPhase containing base game configuration
//   - error if reading or parsing fails
func readBasePhase(
	csvReader *reader.Adapter, modPath *tools.PathBuilder, general *configs.GeneralConfig,
) (*base.SpinPhase, error) {
	const (
		basePrefix = "base"
	)

	// PathBuilder for creating paths to files
	phasePath := modPath.Phase(basePrefix)

	dist, err := common.ReadDistributions(csvReader, phasePath)
	if err != nil {
		return nil, fmt.Errorf("read distributions: %w", err)
	}

	gen, err := readBaseReelsGenerator(csvReader, phasePath, general, dist.CopyValues())
	if err != nil {
		return nil, fmt.Errorf("read base reel generator: %w", err)
	}

	waysCalculator := readWaysCalculator(general)

	spinPhase := base.NewSpinPhase(dist, gen, waysCalculator, general.FsInitCounter)

	return spinPhase, nil
}

// readWaysCalculator constructs the line-win calculator from the shared general
// configuration. Wild and non-scoring symbols (coin, collector) are registered
// so the evaluator handles substitutions and skips correctly.
//
// Parameters:
//   - general: shared game configuration supplying paytable, win lines, and symbol IDs
//
// Returns:
//   - pointer to Calculator ready to evaluate line wins on base-game boards
func readWaysCalculator(
	general *configs.GeneralConfig,
) calculator.Calculators {
	wilds := make(model.Collection, 0, len(general.WildIDs))
	wilds = append(wilds, general.WildIDs...)

	skipSymbol := model.Collection{rules.ScatterID}

	waysCalculators := calculators.NewWays(general.PayTable, wilds, skipSymbol)

	return calculator.NewCalculator(waysCalculators)
}

// readBaseReelsGenerator reads reel-strip data for every trigger group and
// builds the standard board generator used for base-game spins.
//
// Parameters:
//   - csvReader: CSV file adapter backed by the embedded config FS
//   - phasePath: path builder pointing to the base-game phase directory
//   - general: shared game configuration supplying the per-reel symbol counts
//   - groupIDs: trigger group IDs for which reel files must be loaded
//
// Returns:
//   - pointer to Generator with all group reel strips loaded
//   - error if any reel file cannot be read or the generator cannot be initialised
func readBaseReelsGenerator(
	csvReader *reader.Adapter,
	phasePath *tools.PathBuilder,
	general *configs.GeneralConfig,
	groupIDs []model.GroupID,
) (*generator.Generator, error) {
	reels, err := common.ReadReels(csvReader, phasePath, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("read reels: %w", err)
	}

	reelGen, err := generator.NewReelsGenerator(reels, general.BaseSymbolsOnReels)
	if err != nil {
		return nil, fmt.Errorf("make base board generator: %w", err)
	}

	return reelGen, nil
}
