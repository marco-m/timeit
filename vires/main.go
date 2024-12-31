// This file is the vis entrypoint.

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/marco-m/clim"
	"github.com/marco-m/vis"
)

// HOW TO USE
// 1. cd to the project root directory
// 2. ./vis -h

func main() {
	os.Exit(mainInt())
}

func mainInt() int {
	err := mainErr(os.Args[1:])
	if err == nil {
		return 0
	}
	if errors.Is(err, clim.ErrHelp) {
		fmt.Println(err)
		return 0
	}
	fmt.Println(err)
	return 1
}

type App struct{}

func mainErr(args []string) error {
	app := App{}
	cli, err := clim.NewTop[App]("vis", "the build system of the timeit project", nil)
	if err != nil {
		return err
	}

	if err := newCleanCmd(cli); err != nil {
		return err
	}
	if err := newBuildCmd(cli); err != nil {
		return err
	}
	if err := newTestCmd(cli); err != nil {
		return err
	}

	action, err := cli.Parse(args)
	if err != nil {
		return err
	}

	// Ignore SIGINT, allowing subprocesses time to properly clean up.
	vis.ConsumeSignals()

	// TODO
	// Here is a good place to check for the existence of the tools that we will need:
	// - gopass
	// - terraform
	// - packer
	// - ...

	return action(app)
}
