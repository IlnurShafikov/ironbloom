package wild

import "errors"

// ErrInvalidWildID is returned when a wild symbol has no configured multiplier.
var ErrInvalidWildID = errors.New("invalid wild id")
