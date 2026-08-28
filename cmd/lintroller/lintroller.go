// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getoutreach/gobox/pkg/events"
	"github.com/getoutreach/gobox/pkg/log"
	"github.com/getoutreach/lintroller/internal/config"
	"github.com/getoutreach/lintroller/internal/copyright"
	"github.com/getoutreach/lintroller/internal/doculint"
	"github.com/getoutreach/lintroller/internal/errorlint"
	"github.com/getoutreach/lintroller/internal/header"
	"github.com/getoutreach/lintroller/internal/todo"
	"github.com/getoutreach/lintroller/internal/why"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	const configHelp = "the path to the config file for lintroller. " +
		"If this is not set it will be assumed lintroller is running as a vet tool."
	const quietHelp = "if set, emit log statements outside of linting results. " +
		"Only applies when config is given."
	// This needs to be set so that when the analyzers parse their flags they won't error due to
	// an unknown flag being passed.
	_ = flag.String("config", "", configHelp)
	_ = flag.Bool("quiet", true, quietHelp)

	mainFs := flag.NewFlagSet("lintroller", flag.ContinueOnError)

	var configPath string
	var quiet bool

	mainFs.StringVar(&configPath, "config", "", configHelp)
	mainFs.BoolVar(&quiet, "quiet", true, quietHelp)

	_ = mainFs.Parse(os.Args[1:]) //nolint:errcheck // Why: There is no need to check this error.

	if configPath != "" {
		if quiet {
			log.SetOutput(io.Discard)
		}

		cfg, err := config.FromFile(configPath)
		if err != nil {
			log.Fatal(context.Background(), "retrieve config from file", events.NewErrorInfo(err))
		}

		log.Info(context.Background(), "config gathered from file", cfg, log.F{
			"path": configPath,
		})

		exclusionPaths, err := cfg.Lintroller.CompiledExclusionPaths()
		if err != nil {
			log.Fatal(context.Background(), "compile exclusion paths", events.NewErrorInfo(err))
		}

		table := []struct {
			Enabled  bool
			Analyzer *analysis.Analyzer
		}{
			{cfg.Header.Enabled, header.NewAnalyzerWithOptions(strings.Join(cfg.Header.Fields, ","), cfg.Header.Warn)},
			{cfg.Copyright.Enabled, copyright.NewAnalyzerWithOptions(cfg.Copyright.Text, cfg.Copyright.Pattern, cfg.Copyright.Warn)},
			{cfg.Doculint.Enabled, doculint.NewAnalyzerWithOptions(cfg.Doculint.MinFunLen,
				cfg.Doculint.ValidatePackages, cfg.Doculint.ValidateFunctions, cfg.Doculint.ValidateVariables,
				cfg.Doculint.ValidateConstants, cfg.Doculint.ValidateTypes, cfg.Doculint.Warn)},
			{cfg.Todo.Enabled, todo.NewAnalyzerWithOptions(cfg.Todo.Warn)},
			{cfg.Why.Enabled, why.NewAnalyzerWithOptions(cfg.Why.Warn)},
		}

		var analyzers []*analysis.Analyzer
		for i := range table {
			if table[i].Enabled {
				analyzers = append(analyzers, withExclusionPaths(table[i].Analyzer, exclusionPaths))
			}
		}

		if len(analyzers) > 0 {
			multichecker.Main(analyzers...)
		}
		return
	}

	unitchecker.Main(
		&doculint.Analyzer,
		&header.Analyzer,
		&copyright.Analyzer,
		&todo.Analyzer,
		&why.Analyzer,
		&errorlint.Analyzer,
	)
}

// withExclusionPaths wraps the given analyzer and skips execution for excluded paths.
func withExclusionPaths(analyzer *analysis.Analyzer, exclusionPaths []*regexp.Regexp) *analysis.Analyzer {
	if len(exclusionPaths) == 0 {
		return analyzer
	}

	wrapped := *analyzer
	wrapped.Run = func(pass *analysis.Pass) (interface{}, error) {
		if isExcludedPass(pass, exclusionPaths) {
			return nil, nil
		}

		return analyzer.Run(pass)
	}

	return &wrapped
}

// isExcludedPass reports whether the analysis pass should be skipped due to exclusions.
func isExcludedPass(pass *analysis.Pass, exclusionPaths []*regexp.Regexp) bool {
	if matchesAnyExclusion(pass.Pkg.Path(), exclusionPaths) {
		return true
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if matchesAnyExclusion(filename, exclusionPaths) {
			return true
		}
	}

	return false
}

// matchesAnyExclusion reports whether a path matches any configured exclusion pattern.
func matchesAnyExclusion(path string, exclusionPaths []*regexp.Regexp) bool {
	normalizedPath := strings.ReplaceAll(filepath.ToSlash(path), "\\", "/")
	for _, exclusionPath := range exclusionPaths {
		if exclusionPath.MatchString(normalizedPath) {
			return true
		}
	}

	return false
}
