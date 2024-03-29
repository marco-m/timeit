// This code is released under the MIT License
// Copyright (c) 2023 Marco Molteni and the timeit contributors.

package timeit_test

import (
	"os"
	"testing"

	"github.com/marco-m/timeit/pkg/pytestsim"
	"github.com/marco-m/timeit/pkg/sleepit"
	"github.com/marco-m/timeit/pkg/timeit"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"timeit":    timeit.Main,
		"sleepit":   sleepit.Main,
		"pytestsim": pytestsim.Main,
	}))
}

// NOTE Since kong, used by timeit, calls os.Exit directly, we miss some coverage
// information :-(
func TestScriptTimeit(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
