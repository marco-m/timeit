package main

import (
	"fmt"

	"github.com/marco-m/clim"
	"github.com/marco-m/vis"
)

type testCmd struct {
	Cover   bool
	Target  string
	Browser bool
}

func newTestCmd(parent *clim.CLI[App]) error {
	testCmd := testCmd{}

	cli, err := clim.NewSub(parent, "test", "run the tests", testCmd.Run)
	if err != nil {
		return err
	}

	if err := cli.AddFlags(
		&clim.Flag{
			Value: clim.Bool(&testCmd.Cover, false),
			Long:  "cover", Help: "Calculate code coverage",
		},
		&clim.Flag{
			Value: clim.String(&testCmd.Target, "all"),
			Short: "t", Long: "target",
			Help: "Test target (one of: all unit testscript)",
		},
		&clim.Flag{
			Value: clim.Bool(&testCmd.Browser, false),
			Long:  "browser", Help: "Open browser on code coverage results (needs --cover)",
		},
	); err != nil {
		return err
	}

	return nil
}

// TODO: implement slowest tests, see
// https://github.com/gotestyourself/gotestsum?tab=readme-ov-file#finding-and-skipping-slow-tests
func (cmd *testCmd) Run(app App) error {
	const coverfile = "out/cover.out"
	if cmd.Browser && !cmd.Cover {
		return clim.NewParseError("--browser requires --cover")
	}

	var args []string
	switch cmd.Target {
	case "all":
		// do nothing
	case "unit":
		args = append(args, "-skip=^TestScript")
	case "testscript":
		args = append(args, "-run=^TestScript")
	default:
		return clim.NewParseError("unknown test target: %s", cmd.Target)
	}
	if cmd.Cover {
		args = append(args, "-coverprofile", coverfile)
	}
	args = append(args, "./...")

	if err := vis.GoTestSum(args...); err != nil {
		return fmt.Errorf("test: failures")
	}
	if cmd.Browser {
		vis.GoCoverageBrowser(coverfile)
	}

	return nil
}
