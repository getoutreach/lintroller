package main

import (
	"context"
	"flag"
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

func main() {
	// This needs to be set so that when the analyzers parse their flags they won't error due to
	// an unknown flag being passed.
	_ = flag.String("config", "", "The path to the config file for lintroller. If this is not set it will be assumed lintroller is running as a vet tool.")

	mainFs := flag.NewFlagSet("main", flag.ContinueOnError)

	var configPath string
	mainFs.StringVar(&configPath, "config", "", "The path to the config file for lintroller. If this is not set it will be assumed lintroller is running as a vet tool.")
	_ = mainFs.Parse(os.Args[1:]) //nolint:errcheck // Why: There is no need to check this error.

	if configPath != "" {
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
			{cfg.Doculint.Enabled, doculint.NewAnalyzerWithOptions(cfg.Doculint.MinFunLen, cfg.Doculint.ValidatePackages, cfg.Doculint.ValidateFunctions, cfg.Doculint.ValidateVariables, cfg.Doculint.ValidateConstants, cfg.Doculint.ValidateTypes)},
			{cfg.Todo.Enabled, &todo.Analyzer},
			{cfg.Why.Enabled, &why.Analyzer},
		}

		var analyzers []*analysis.Analyzer
		for i := range table {
			if table[i].Enabled {
				analyzers = append(analyzers, table[i].Analyzer)
			}
		}

		multichecker.Main(analyzers...)
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
