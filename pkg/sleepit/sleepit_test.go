// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package sleepit_test

import (
	"os"
	"testing"

	"github.com/marco-m/timeit/pkg/sleepit"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"sleepit": sleepit.Main,
	}))
}

func TestScriptSleepit(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
