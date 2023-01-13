// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: This file implements functional options for the Pass type.

package reporter

// PassOption is the functional argument type for Pass.
type PassOption func(*Pass)

// Warn will ensure that all reported lint issues are warnings as opposed to errors
// for the current linter.
func Warn() PassOption {
	return func(p *Pass) {
		p.warn = true
	}
}
