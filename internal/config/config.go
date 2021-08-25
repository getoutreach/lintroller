// Package config is used for defining the configuration necessary to run the
// linters in an automated fashion.
package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is parent type we use to unmarshal YAML files into to gather config
// for the lintroller.
type Config struct {
	Lintroller `yaml:"lintroller"`
}

// MarshalLog implements the log.Marshaler interface.
func (c *Config) MarshalLog(addField func(key string, value interface{})) {
	addField("lintroller", c.Lintroller)
}

// FromFile decodes a Config type given a file path.
func FromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open config file")
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "decode config file")
	}

	if err := cfg.Lintroller.ValidateTier(); err != nil {
		return nil, errors.Wrap(err, "validate the tier given to lintroller")
	}

	return &cfg, nil
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

// MarshalLog implements the log.Marshaler interface.
func (lr *Lintroller) MarshalLog(addField func(key string, value interface{})) {
	addField("header", lr.Header)
	addField("copyright", lr.Copyright)
	addField("doculint", lr.Doculint)
	addField("todo", lr.Todo)
	addField("why", lr.Why)
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

// MarshalLog implements the log.Marshaler interface.
func (h *Header) MarshalLog(addField func(key string, value interface{})) {
	addField("enabled", h.Enabled)
	addField("fields", h.Fields)
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

// MarshalLog implements the log.Marshaler interface.
func (c *Copyright) MarshalLog(addField func(key string, value interface{})) {
	addField("enabled", c.Enabled)
	addField("string", c.String)
	addField("regex", c.Regex)
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

// MarshalLog implements the log.Marshaler interface.
func (d *Doculint) MarshalLog(addField func(key string, value interface{})) {
	addField("enabled", d.Enabled)
	addField("minFunLen", d.MinFunLen)
	addField("validatePackages", d.ValidatePackages)
	addField("validateFunctions", d.ValidateFunctions)
	addField("validateVariables", d.ValidateVariables)
	addField("validateConstants", d.ValidateConstants)
	addField("validateTypes", d.ValidateTypes)
}

// Todo is the configuration type that matches the flags exposed by the todo linter.
type Todo struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`
}

// MarshalLog implements the log.Marshaler interface.
func (t *Todo) MarshalLog(addField func(key string, value interface{})) {
	addField("enabled", t.Enabled)
}

// Why is the configuration type that matches the flags exposed by the why linter.
type Why struct {
	// Enabled denotes whether or not this linter is enabled. Defaults to true.
	Enabled bool `yaml:"enabled"`
}

// MarshalLog implements the log.Marshaler interface.
func (w *Why) MarshalLog(addField func(key string, value interface{})) {
	addField("enabled", w.Enabled)
}
