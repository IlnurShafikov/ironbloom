package common

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/configs"
	"github.com/IlnurShafikov/ironbloom/internal/entities/pairs"
	"github.com/IlnurShafikov/ironbloom/internal/tools"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/reader"
)

// ReadMultipliers reads and parses multiplier configurations from CSV files.
// It loads the maximum win multiplier and cost multiplier for the simple game mode.
//
// Parameters:
//   - reader: CSV file adapter for reading configuration files
//   - generalPath: path builder pointing to the general configuration directory
//
// Returns:
//   - pointer to MultipliersConfigs containing all multiplier values
//   - error if reading or parsing fails
func ReadMultipliers(
	csvReader *reader.Adapter,
	generalPath *tools.PathBuilder,
) (*configs.MultipliersConfigs, error) {
	const (
		maxWinName        = "max_win_limit"
		simpleCostName    = "simple_cost_multiplier"
		wildMultiPairName = "wild_multipliers"
	)

	maxMultiplier, err := csvReader.ReadUint64(generalPath.File(maxWinName))
	if err != nil {
		return nil, fmt.Errorf("read max win multiplier: %w", err)
	}

	baseCostMultiplier, err := csvReader.ReadUint64(generalPath.File(simpleCostName))
	if err != nil {
		return nil, fmt.Errorf("read base cost multiplier: %w", err)
	}

	wildMulti, err := readWildMultipliers(csvReader, generalPath)
	if err != nil {
		return nil, fmt.Errorf("read wild multiplier: %w", err)
	}

	return &configs.MultipliersConfigs{
		MaxWinMultiplier:     maxMultiplier,
		SimpleCostMultiplier: baseCostMultiplier,
		WildMultipliers:      wildMulti,
	}, nil
}

func readWildMultipliers(
	csvReader *reader.Adapter,
	generalPath *tools.PathBuilder,
) (pairs.Pairs, error) {
	const wildMultiPairName = "wild_multipliers"

	data, err := csvReader.ReadCSV(generalPath.File(wildMultiPairName))
	if err != nil {
		return nil, fmt.Errorf("read wild multipliers data: %w", err)
	}

	resp, err := pairs.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse wild multipliers pairs: %w", err)
	}

	return resp, nil
}
