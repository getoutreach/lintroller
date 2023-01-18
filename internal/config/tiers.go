// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains tiering information that links back to certain sets
// of configuration. This is really just to help ease the rollout of custom lint rules.
// This set of tiers is subject to change, but efforts will be made to never "move the
// goal posts" of already defined tiers.

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/getoutreach/gobox/pkg/log"
	"github.com/pkg/errors"
)

// Configuration tier variables that correspond to ops-level tiers. The reason these
// are variables as opposed to constants is because the address of them are needed
// when defining their corresponding Tier*Configuration types.
var (
	// TierBronze is the tier name that corresponds to the TierBronzeConfiguration
	// configuration minimums.
	TierBronze = "bronze"

	// TierSilver is the tier name that corresponds to the TierSilverConfiguration
	// configuration minimums.
	TierSilver = "silver"

	// TierGold is the tier name that corresponds to the TierGoldConfiguration
	// configuration minimums.
	TierGold = "gold"

	// TierPlatinum is the tier name that corresponds to the TierPlatinumConfiguration
	// configuration minimums.
	TierPlatinum = "platinum"
)

// ValidateTier ensures that if a tier was provided, the rest of the configuration
// meets minimum requirements. If a field was left unset, it will be automatically
// set to the minimum requirement.
func (l *Lintroller) ValidateTier() error {
	if l.Tier == nil {
		// No tier selected, nothing to validate.
		return nil
	}

	switch strings.ToLower(*l.Tier) {
	case TierBronze:
		if err := l.EnsureMinimums(&TierBronzeConfiguration); err != nil {
			return errors.Wrap(err, "ensure given configuration meets minimum requirements for bronze tier")
		}
	case TierSilver:
		if err := l.EnsureMinimums(&TierSilverConfiguration); err != nil {
			return errors.Wrap(err, "ensure given configuration meets minimum requirements for silver tier")
		}
	case TierGold:
		if err := l.EnsureMinimums(&TierGoldConfiguration); err != nil {
			return errors.Wrap(err, "ensure given configuration meets minimum requirements for gold tier")
		}
	case TierPlatinum:
		if err := l.EnsureMinimums(&TierPlatinumConfiguration); err != nil {
			return errors.Wrap(err, "ensure given configuration meets minimum requirements for platinum tier")
		}
	default:
		log.Warn(context.Background(),
			fmt.Sprintf("provided does not match any of the following: %q, %q, %q, %q (sans-quotes)",
				TierBronze, TierSilver, TierGold, TierPlatinum),
			log.F{
				"tier": *l.Tier,
			})
	}

	return nil
}

// EnsureMinimums takes a desired Lintroller variable and diffs it against the receiver. It
// will automatically override booleans set to false, needing to be set to true, as well as
// any zero-valued struct field.
//
// This function will allow the receiver to be more restrictive (enable linters when the
// desired has them disabled, set the minimum function length to a lower value, add more
// required header fields, etc.), but not allow it to be less restrictive.
func (l *Lintroller) EnsureMinimums(desired *Lintroller) error { //nolint:funlen // Why: Splitting this function out would add no value.
	overrideBool := func(necessary, current bool, fieldPath string) bool {
		if necessary {
			if !current {
				// If the necessary is true, but the current is false, then override it and log this action.
				log.Warn(context.Background(),
					"boolean value required to be true to meet tier minimum stanards is set to false - overriding to true",
					log.F{
						"field": fieldPath,
					})
				return true
			}
		}

		return current
	}

	// Ensure header linter minimum configuration against desired.
	l.Header.Enabled = overrideBool(desired.Header.Enabled, l.Header.Enabled, "lintroller.header.enabled")
	if l.Header.Enabled {
		if l.Header.Fields == nil {
			l.Header.Fields = desired.Header.Fields
		} else {
			for i := range desired.Header.Fields {
				var found bool
				for j := range l.Header.Fields {
					if desired.Header.Fields[i] == l.Header.Fields[j] {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf(
						"deviation detected from tier minimum defaults in lintroller.header.fields, fields must contain \"%s\"",
						desired.Header.Fields[i])
				}
			}
		}
	}

	// Ensure copyright linter minimum configuration against desired.
	l.Copyright.Enabled = overrideBool(desired.Copyright.Enabled, l.Copyright.Enabled, "lintroller.copyright.enabled")
	if l.Copyright.Pattern == "" {
		log.Warn(context.Background(), "zero value detected for field, overriding to value found in desired tier minimum version", log.F{
			"field": "lintroller.copyright.pattern",
			"value": desired.Copyright.Pattern,
		})

		l.Copyright.Pattern = desired.Copyright.Pattern
	} else if l.Copyright.Pattern != desired.Copyright.Pattern {
		log.Warn(context.Background(), "deviation detected for field, overriding to value found in desired tier minimum version", log.F{
			"field": "lintroller.copyright.pattern",
			"value": desired.Copyright.Pattern,
		})

		l.Copyright.Pattern = desired.Copyright.Pattern
	}

	// Ensure doculint linter minimum configuration against desired.
	l.Doculint.Enabled = overrideBool(desired.Doculint.Enabled, l.Doculint.Enabled, "lintroller.doculint.enabled")
	if l.Doculint.Enabled {
		l.Doculint.ValidatePackages = overrideBool(
			desired.Doculint.ValidatePackages, l.Doculint.ValidatePackages, "lintroller.doculint.validatePackages")
		l.Doculint.ValidateFunctions = overrideBool(
			desired.Doculint.ValidateFunctions, l.Doculint.ValidateFunctions, "lintroller.doculint.validateFunctions")
		l.Doculint.ValidateVariables = overrideBool(
			desired.Doculint.ValidateVariables, l.Doculint.ValidateVariables, "lintroller.doculint.validateVariables")
		l.Doculint.ValidateConstants = overrideBool(
			desired.Doculint.ValidateConstants, l.Doculint.ValidateConstants, "lintroller.doculint.validateConstants")
		l.Doculint.ValidateTypes = overrideBool(
			desired.Doculint.ValidateTypes, l.Doculint.ValidateTypes, "lintroller.doculint.validateTypes")

		if l.Doculint.ValidateFunctions {
			if l.Doculint.MinFunLen == 0 {
				log.Warn(context.Background(), "zero value detected for field, overriding to value found in desired tier minimum version", log.F{
					"field": "lintroller.doculint.minFunLen",
					"value": desired.Doculint.MinFunLen,
				})
				l.Doculint.MinFunLen = desired.Doculint.MinFunLen
			} else if l.Doculint.MinFunLen > desired.Doculint.MinFunLen || l.Doculint.MinFunLen < 0 {
				return fmt.Errorf(
					"deviation detected from tier minimum defaults in lintroller.doculint.minfunlen, minfunlen must be set within (0, %d]",
					desired.Doculint.MinFunLen)
			}
		}
	}

	// Ensure todo linter minimum configuration against desired.
	l.Todo.Enabled = overrideBool(desired.Todo.Enabled, l.Todo.Enabled, "lintroller.todo.enabled")

	// Ensure why linter minimum configuration against desired.
	l.Why.Enabled = overrideBool(desired.Todo.Enabled, l.Todo.Enabled, "lintroller.why.enabled")

	return nil
}

// TierBronzeConfiguration is the Lintroller configuration minumums that correspond
// to the Bronze OpsLevel tier.
var TierBronzeConfiguration = Lintroller{
	Tier: &TierBronze,
	Header: Header{
		Enabled: false,
		Fields:  nil,
	},
	Copyright: Copyright{
		Enabled: false,
	},
	Doculint: Doculint{
		Enabled:           false,
		MinFunLen:         0,
		ValidatePackages:  false,
		ValidateFunctions: false,
		ValidateVariables: false,
		ValidateConstants: false,
		ValidateTypes:     false,
	},
	Todo: Todo{
		Enabled: false,
	},
	Why: Why{
		Enabled: false,
	},
	Errorlint: Errorlint{
		Warn: false,
	},
}

// TierSilverConfiguration is the Lintroller configuration minumums that correspond
// to the Silver OpsLevel tier.
var TierSilverConfiguration = Lintroller{
	Tier: &TierSilver,
	Header: Header{
		Enabled: true,
		Fields:  []string{"Description"},
	},
	Copyright: Copyright{
		Enabled: true,
		Pattern: `^Copyright 20.*$`,
	},
	Doculint: Doculint{
		Enabled:           true,
		MinFunLen:         0,
		ValidatePackages:  true,
		ValidateFunctions: false,
		ValidateVariables: false,
		ValidateConstants: false,
		ValidateTypes:     false,
	},
	Todo: Todo{
		Enabled: true,
	},
	Why: Why{
		Enabled: true,
	},
	Errorlint: Errorlint{
		Warn: false,
	},
}

// TierGoldConfiguration is the Lintroller configuration minumums that correspond
// to the Gold OpsLevel tier.
var TierGoldConfiguration = Lintroller{
	Tier: &TierSilver,
	Header: Header{
		Enabled: true,
		Fields:  []string{"Description"},
	},
	Copyright: Copyright{
		Enabled: true,
		Pattern: `^Copyright 20.*$`,
	},
	Doculint: Doculint{
		Enabled:           true,
		MinFunLen:         0,
		ValidatePackages:  true,
		ValidateFunctions: false,
		ValidateVariables: true,
		ValidateConstants: true,
		ValidateTypes:     true,
	},
	Todo: Todo{
		Enabled: true,
	},
	Why: Why{
		Enabled: true,
	},
	Errorlint: Errorlint{
		Warn: true,
	},
}

// TierPlatinumConfiguration is the Lintroller configuration minumums that correspond
// to the Platinum OpsLevel tier.
var TierPlatinumConfiguration = Lintroller{
	Tier: &TierSilver,
	Header: Header{
		Enabled: true,
		Fields:  []string{"Description"},
	},
	Copyright: Copyright{
		Enabled: true,
		Pattern: `^Copyright 20.*$`,
	},
	Doculint: Doculint{
		Enabled:           true,
		MinFunLen:         10,
		ValidatePackages:  true,
		ValidateFunctions: true,
		ValidateVariables: true,
		ValidateConstants: true,
		ValidateTypes:     true,
	},
	Todo: Todo{
		Enabled: true,
	},
	Why: Why{
		Enabled: true,
	},
	Errorlint: Errorlint{
		Warn: true,
	},
}
