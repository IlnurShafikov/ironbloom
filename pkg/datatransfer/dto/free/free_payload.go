package free

import (
	"github.com/IlnurShafikov/ironbloom/internal/entities/result/events"
	"github.com/IlnurShafikov/ironbloom/pkg/datatransfer/dto/common"
)

// Payload contains the detailed game data for a free spin step.
type Payload struct {
	// Events contains all game events that occurred (wins, retriggers, etc.)
	Events []events.Event `json:"events,omitempty"`

	// Symbols represents the initial symbol board before any transformations
	Symbols []*common.Symbol `json:"symbols"`

	// Counter tracks the free spins (added, left, played)
	Counter common.Counter `json:"counter"`

	// NextStepType indicates what type of step comes next (free, respin, or base)
	NextStepType int `json:"nextStepType"`
}
