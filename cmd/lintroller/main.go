package main

import (
	"github.com/getoutreach/lintroller/internal/copyright"
	"github.com/getoutreach/lintroller/internal/doculint"
	"github.com/getoutreach/lintroller/internal/header"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(
		&doculint.Analyzer,
		&header.Analyzer,
		&copyright.Analyzer,
		// Add more *analysis.Analyzer's here.
	)
}
