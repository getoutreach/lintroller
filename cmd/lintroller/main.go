package main

import (
	"github.com/getoutreach/lintroller/internal/doculint"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(
		&doculint.Analyzer,
		// Add more *analysis.Analyzer's here.
	)
}
