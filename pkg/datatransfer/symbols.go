package datatransfer

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/result"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/common"
	"github.com/IlnurShafikov/slot-engine/pkg/model"
)

// TransformSymbols converts the internal game board representation into client DTO symbols.
// This function should be implemented to transform your game's board structure into
// an array of Symbol DTOs that will be sent to the client.
//
// Parameters:
//   - gameBoard: the internal game board (type depends on your implementation)
//
// Returns:
//   - slice of Symbol DTOs representing the board
//   - error if transformation fails
func TransformSymbols(
	gameBoard model.Board, wildMulti []result.WildMultiplier,
) ([]*common.Symbol, error) {
	if gameBoard == nil {
		return nil, ErrGameBoardIsNil
	}

	resp := transformBoardToSymbols(gameBoard)

	updateWildWithMultiplier(wildMulti, resp)

	return resp, nil
}

// transformBoardToSymbols converts the board result into a flat array of Symbol DTOs.
// Each symbol includes its ID, position on the board, and optional properties.
//
// Parameters:
//   - boardResult: the board result containing the symbol layout
//
// Returns:
//   - slice of Symbol DTOs in row-major order (reel by reel, then row by row)
func transformBoardToSymbols(gameBoard model.Board) []*common.Symbol {
	reelsCount := gameBoard.ReelCount()
	rowsCount := gameBoard.FirstReelSize()

	resp := make([]*common.Symbol, 0, reelsCount*rowsCount)

	for reel := range reelsCount {
		for row := range rowsCount {
			resp = append(resp, &common.Symbol{
				SymbolID: gameBoard[reel][row],
				Position: model.NewPosition(reel, row),
				Property: nil,
			})
		}
	}

	return resp
}

// updateWildWithMultiplier attaches the multiplier property to the client
// symbols that stand on the positions of the wild symbols.
//
// Parameters:
//   - wildMulti: wild symbols with their multipliers
//   - symbols: client symbols to update in place
func updateWildWithMultiplier(
	wildMulti []result.WildMultiplier,
	symbols []*common.Symbol,
) {
	if len(wildMulti) == 0 {
		return
	}

	multiByPosition := make(map[model.Position]result.WildMultiplier, len(wildMulti))

	for _, wildMulti := range wildMulti {
		multiByPosition[wildMulti.Position()] = wildMulti
	}

	for i, symbol := range symbols {
		multi, ok := multiByPosition[symbol.Position]
		if ok {
			symbols[i].Property = &common.SymbolProperties{
				Multipier: multi.MultiplierValue(),
			}
		}
	}
}
