// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/marco-m/taschino/pkg/release"
)

var (
	// Filled by the linker.
	fullVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

func main() {
	os.Exit(run("timeit", os.Args[1:], os.Stderr, nil))
}

// Test escape hatch.
var execCommand = exec.Command

func run(progname string, args []string, out io.Writer, started chan<- (struct{})) int {
	flag.CommandLine.SetOutput(out)
	flagSet := flag.NewFlagSet(progname, flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(out, "%s measures the time of command execution\n\n", progname)
		fmt.Fprintf(out, "Usage: %s command\n\n", progname)
		fmt.Fprintf(out, "Options:\n")
		flagSet.PrintDefaults()
	}

	var (
		showVersion  = flagSet.Bool("version", false, "show version")
		checkVersion = flagSet.Bool("check-version", false, "check online if new version is available")
	)

	if flagSet.Parse(args) != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(out, "timeit version", fullVersion)
		return 0
	}
	if *checkVersion {
		human_url := fmt.Sprintf("https://github.com/%s/%s", "marco-m", "timeit")
		latestVersion, err := release.GitHubLatest("marco-m", "timeit")
		if err != nil {
			fmt.Fprintln(out, err)
			return 1
		}
		result, err := release.Compare(shortVersion, latestVersion)
		if err != nil {
			fmt.Fprintln(out, err)
			return 1
		}
		switch result {
		case 0:
			fmt.Fprintf(out, "installed version %s is the same as the latest version %s\n",
				shortVersion, latestVersion)
		case -1:
			fmt.Fprintf(out, "installed version %s is older than the latest version %s\n",
				shortVersion, latestVersion)
			fmt.Fprintln(out, "To upgrade visit", human_url)
		case +1:
			fmt.Fprintf(out, "(unexpected?) installed version %s is newer than the latest version %s\n",
				shortVersion, latestVersion)
		}
		return 0
	}

	if len(flagSet.Args()) == 0 {
		fmt.Fprintln(out, "timeit: expected a command to time")
		return 2
	}

	cmd := execCommand(flagSet.Args()[0], flagSet.Args()[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	if startErr := cmd.Start(); startErr != nil {
		fmt.Fprintln(out, "timeit: start child:", startErr)
		return 1
	}

	// We are in the parent, after having started the child.
	// Ignoring SIGINT as the original /usr/bin/time does with
	// signal.Ignore(os.Interrupt) has subtle side-effects with the tests.
	// Thus, we do the equivalent with a do-nothing signal handler.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
	}()

	if started != nil {
		go func() {
			started <- struct{}{}
		}()
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		fmt.Fprintln(out, "timeit: wait child:", waitErr)
	}
	fmt.Fprintf(out,
		`timeit results:
real: %v
`,
		time.Since(start))
	code := cmd.ProcessState.ExitCode()
	if code == -1 {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		// Follow the shell convention, https://en.wikipedia.org/wiki/Exit_status
		code = 128 + int(status.Signal())
	}
	return code
}
