package main_test

import (
	"os"
	"testing"

	"github.com/marco-m/timeit/cmd/understand-script"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"understand-script": main.Main,
	}))
}

// NOTE Since understand-script calls os.Exit directly, we miss some coverage information :-(
func TestScriptUnderstandScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
