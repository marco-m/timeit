package main

import (
	"fmt"
	"strings"

	"github.com/marco-m/clim"
	"github.com/marco-m/vis"
)

// Treat this as a constant!
var kArtifacts = []string{
	"timeit",
	"sleepit",
	"understand-script",
	"pytestsim",
}

type CleanCmd struct{}

func newCleanCmd(parent *clim.CLI[App]) error {
	cleanCmd := CleanCmd{}

	_, err := clim.NewSub(parent, "clean", "clean build artifacts", cleanCmd.Run)
	if err != nil {
		return err
	}
	return nil
}

func (cmd CleanCmd) Run(app App) error {
	vis.RemoveAllFiles(kArtifacts...)
	return nil
}

type buildCmd struct {
	Goos    string
	Targets []string
}

func newBuildCmd(parent *clim.CLI[App]) error {
	buildCmd := buildCmd{}

	cli, err := clim.NewSub(parent, "build", "build the project", buildCmd.Run)
	if err != nil {
		return err
	}

	if err := cli.AddFlags(
		&clim.Flag{
			Value: clim.String(&buildCmd.Goos, ""),
			Long:  "goos", Help: "compilation target OS (default: same as host)",
		},
		&clim.Flag{
			Value: clim.StringSlice(&buildCmd.Targets, kArtifacts),
			Short: "t", Long: "targets",
			Help: "comma-separated list of build targets",
		},
	); err != nil {
		return err
	}

	return nil
}

func (cmd *buildCmd) Run(app App) error {
	shortVersion, err := vis.GitShortVersion()
	if err != nil {
		return err
	}
	longVersion, err := vis.GitLongVersion()
	if err != nil {
		return err
	}
	ldflags := strings.Join([]string{
		"-X github.com/marco-m/timeit/pkg/timeit.longVersion=" + longVersion,
		"-X github.com/marco-m/timeit/pkg/timeit.shortVersion=" + shortVersion,
	}, " ")
	// NOTE We don't look at the error type; this can hide useful information ...
	failCount := 0
	for _, tgt := range cmd.Targets {
		// FIXME if I pass ldflags as quoted, go build returns an error.
		// If I don't, "go build" is happy, but the log output misses the quotes,
		// so it cannot be copy and pasted :-(
		pkg := vis.FilepathJoinDot("cmd", tgt)
		args := []string{"-ldflags", ldflags, pkg}
		// usage: go build [-o output] [build flags] [packages]
		if err := vis.GoBuild(cmd.Goos, args...); err != nil {
			failCount++
		}
	}
	if failCount > 0 {
		return fmt.Errorf("build: %d failures", failCount)
	}
	return nil
}
