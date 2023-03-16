package main

import (
	"github.com/speed1313/enumResearch"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(enumResearch.Analyzer) }
