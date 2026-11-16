package readers

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/configs"
	"github.com/IlnurShafikov/ironbloom/internal/configs/readers/common"
	"github.com/IlnurShafikov/ironbloom/internal/tools"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/reader"
)

const (
	// generalPrefix defines the directory name for general configuration files
	generalPrefix = "general"
	// baseSymbolOnReelName is the CSV file name for the per-reel symbol counts in base phase.
	baseSymbolOnReelName = "base_symbols_on_reel"
	// freeSymbolOnReelName is the CSV file name for the per-reel symbol counts in free phase.
	freeSymbolOnReelName = "free_symbols_on_reel"
	// wildIDsName is the CSV file name that lists the wild symbol IDs.
	wildIDsName = "wild_ids"
)

// Read loads and parses the general game configuration.
// It reads the paytable and multipliers from CSV files in the general directory.
//
// Parameters:
//   - csvReader: CSV file adapter for reading configuration files
//
// Returns:
//   - pointer to GeneralConfig containing paytable and multipliers
//   - error if reading or parsing fails
func Read(csvReader *reader.Adapter) (*configs.GeneralConfig, error) {
	generalPath := tools.NewPathBuilder(generalPrefix)

	paytable, err := common.ReadPaytable(csvReader, generalPath)
	if err != nil {
		return nil, fmt.Errorf("read paytable: %w", err)
	}

	baseSymbolsOnReel, err := common.ReadIntSlices(
		csvReader, generalPath.File(baseSymbolOnReelName))
	if err != nil {
		return nil, fmt.Errorf("read base symbols on reel: %w", err)
	}

	freeSymbolsOnReel, err := common.ReadIntSlices(
		csvReader, generalPath.File(freeSymbolOnReelName))
	if err != nil {
		return nil, fmt.Errorf("read free symbols on reel: %w", err)
	}

	wildIDs, err := common.ReadIntSlices(csvReader, generalPath.File(wildIDsName))
	if err != nil {
		return nil, fmt.Errorf("read wild ids: %w", err)
	}

	multipliers, err := common.ReadMultipliers(csvReader, generalPath)
	if err != nil {
		return nil, fmt.Errorf("read multipliers: %w", err)
	}

	fsInitCounter, err := common.ReadFsInitCounter(csvReader, generalPath)
	if err != nil {
		return nil, fmt.Errorf("read fs init counter: %w", err)
	}

	volatility, err := common.ReadVolatility(csvReader, generalPath)
	if err != nil {
		return nil, fmt.Errorf("read volatility: %w", err)
	}

	return &configs.GeneralConfig{
		PayTable:           paytable,
		Multipliers:        multipliers,
		BaseSymbolsOnReels: baseSymbolsOnReel,
		FreeSymbolsOnReels: freeSymbolsOnReel,
		WildIDs:            wildIDs,
		FsInitCounter:      fsInitCounter,
		Volatility:         volatility,
	}, nil
}
