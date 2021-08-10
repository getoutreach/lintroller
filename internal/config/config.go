// Package config is used for defining the configuration necessary to run the
// linters in an automated fashion.
package config

// Config is parent type we use to unmarshal YAML files into to gather config
// for the lintroller.
type Config struct {
	Lintroller `yaml:"lintroller"`
}

// Lintroller contains the actually configuration required by lintroller, used
// by Config. The reason these fields aren't directly in Config is because we
// want to the ability utilize the golangci.yml file for lintroller configuration
// as well, so lintroller configuration needs to be "namespaced" accordingly.
type Lintroller struct {
	// Tier is the desired tier you desire your service to pass for in ops-level.
	Tier *string `yaml:"tier"`

	// Configuration for individual linters proceeding:
	Header    Header    `yaml:"header"`
	Copyright Copyright `yaml:"copyright"`
	Doculint  Doculint  `yaml:"doculint"`
	Todo      Todo      `yaml:"todo"`
	Why       Why       `yaml:"why"`
}

// Header is the configuration type that matches the flags exposed by the header
// linter.
type Header struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`

	// Fields is a list of fields required to be filled out in the header. Defaults
	// to []string{"Description"}.
	Fields []string `yaml:"fields"`
}

// Copyright is the configuration type that matches the flags exposed by the copyright
// linter.
type Copyright struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`

	// String is the copyright string required at the top of each .go file. if empty
	// this linter is a no-op. Defaults to an empty string.
	String string `yaml:"string"`

	// Regex denotes whether or not the copyright string was given as a regular expression.
	// Defaults to false.
	Regex bool `yaml:"regex"`
}

// Doculint is the configuration type that matches the flags exposed by the doculint
// linter.
type Doculint struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`

	// MinFunLen is the minimum function length that doculint will report on if said
	// function has no related documentation. Defaults to 10.
	MinFunLen int `yaml:"minFunLen"`

	// ValidatePackages denotes whether or not package comments should be validated.
	// Defaults to true.
	ValidatePackages bool `yaml:"validatePackages"`

	// ValidateFunctions denotes whether or not function comments should be validated.
	// Defaults to true.
	ValidateFunctions bool `yaml:"validateFunctions"`

	// ValidateVariables denotes whether or not variable comments should be validated.
	// Defaults to true.
	ValidateVariables bool `yaml:"validateVariables"`

	// ValidateConstants denotes whether or not constant comments should be validated.
	// Defaults to true.
	ValidateConstants bool `yaml:"validateConstants"`

	// ValidateTypes denotes whether or not type comments should be validated. Defaults
	// to true.
	ValidateTypes bool `yaml:"validateTypes"`
}

// Todo is the configuration type that matches the flags exposed by the todo linter.
type Todo struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`
}

// Why is the configuration type that matches the flags exposed by the why linter.
type Why struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`
}
