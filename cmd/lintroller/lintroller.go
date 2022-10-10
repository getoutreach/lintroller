// Copyright 2022 Outreach Corporation. All Rights Reserved.

package main

import (
	"context"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/getoutreach/gobox/pkg/events"
	"github.com/getoutreach/gobox/pkg/log"
	"github.com/getoutreach/lintroller/internal/config"
	"github.com/getoutreach/lintroller/internal/copyright"
	"github.com/getoutreach/lintroller/internal/doculint"
	"github.com/getoutreach/lintroller/internal/header"
	"github.com/getoutreach/lintroller/internal/todo"
	"github.com/getoutreach/lintroller/internal/why"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { //nolint:funlen // Why: It wouldn't make sense to split anymore of this into separate functions.
	// This needs to be set so that when the analyzers parse their flags they won't error due to
	// an unknown flag being passed.
	//nolint:lll // Why: Mostly flag info
	_ = flag.String("config", "", "the path to the config file for lintroller. if this is not set it will be assumed lintroller is running as a vet tool")
	//nolint:lll // Why: Mostly flag info
	_ = flag.Bool("quiet", true, "whether or not the linter will emit log statements outside of linting results. only applies when config is passed (when lintroller is not running in vet mode)")

	mainFs := flag.NewFlagSet("main", flag.ContinueOnError)

	var configPath string
	var quiet bool

	//nolint:lll // Why: Mostly flag info
	mainFs.StringVar(&configPath, "config", "", "the path to the config file for lintroller. if this is not set it will be assumed lintroller is running as a vet tool")
	//nolint:lll // Why: Mostly flag info
	mainFs.BoolVar(&quiet, "quiet", true, "whether or not the linter will emit log statements outside of linting results. only applies when config is passed (when lintroller is not running in vet mode)")

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

		table := []struct {
			Enabled  bool
			Analyzer *analysis.Analyzer
		}{
			{cfg.Header.Enabled, header.NewAnalyzerWithOptions(strings.Join(cfg.Header.Fields, ","))},
			{cfg.Copyright.Enabled, copyright.NewAnalyzerWithOptions(cfg.Copyright.Text, cfg.Copyright.Pattern)},
			{cfg.Doculint.Enabled, doculint.NewAnalyzerWithOptions(cfg.Doculint.MinFunLen,
				cfg.Doculint.ValidatePackages, cfg.Doculint.ValidateFunctions, cfg.Doculint.ValidateVariables,
				cfg.Doculint.ValidateConstants, cfg.Doculint.ValidateTypes)},
			{cfg.Todo.Enabled, &todo.Analyzer},
			{cfg.Why.Enabled, &why.Analyzer},
		}

		var analyzers []*analysis.Analyzer
		for i := range table {
			if table[i].Enabled {
				analyzers = append(analyzers, table[i].Analyzer)
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
	)
}
