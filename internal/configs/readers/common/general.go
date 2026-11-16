package common

import (
	"fmt"

	"github.com/IlnurShafikov/ironbloom/internal/tools"
	"github.com/IlnurShafikov/slot-engine/pkg/paytable"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/fscounter"
	"github.com/IlnurShafikov/slot-engine/pkg/tools/reader"
)

// ReadPaytable reads and parses the paytable configuration from a CSV file.
// The paytable defines symbol payouts and count ranges for winning combinations.
//
// Parameters:
//   - csvReader: CSV file adapter for reading configuration files
//   - generalPath: path builder pointing to the general configuration directory
//
// Returns:
//   - pointer to PayTable containing symbol payouts and ranges
//   - error if reading or parsing fails
func ReadPaytable(
	csvReader *reader.Adapter,
	generalPath *tools.PathBuilder,
) (*paytable.PayTable, error) {
	const fileName = "paytable"

	data, err := csvReader.ReadCSV(generalPath.File(fileName))
	if err != nil {
		return nil, fmt.Errorf("read paytable: %w", err)
	}

	resp, err := paytable.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse paytable: %w", err)
	}

	return resp, nil
}

// ReadFsInitCounter reads and parses the free spins initial counter configuration from a CSV file.
// This counter determines the number of free spins awarded at the start.
//
// Parameters:
//   - csvReader: CSV file adapter for reading configuration files
//   - generalPath: path builder pointing to the general configuration directory
//
// Returns:
//   - FSInitCounter containing the initial free spins count distribution
//   - error if reading or parsing fails
func ReadFsInitCounter(
	csvReader *reader.Adapter,
	generalPath *tools.PathBuilder,
) (fscounter.FSInitCounter, error) {
	const fileName = "fs_init_counter"

	data, err := csvReader.ReadCSV(generalPath.File(fileName))
	if err != nil {
		return nil, fmt.Errorf("read data fs init counter: %w", err)
	}

	resp, err := parseDataToSliceInt(data)
	if err != nil {
		return nil, fmt.Errorf("parse data to slice int: %w", err)
	}

	return resp, nil
}
