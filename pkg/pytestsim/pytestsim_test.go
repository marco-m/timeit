// This code is released under the MIT License
// Copyright (c) 2023 Marco Molteni and the timeit contributors.

package pytestsim_test

import (
	"os"
	"testing"

	"github.com/marco-m/timeit/pkg/pytestsim"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"pytestsim": pytestsim.Main,
	}))
}

func TestScriptPytestsim(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
