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

	"github.com/fatih/color"
	"github.com/marco-m/taschino/pkg/release"
	"github.com/mattn/go-isatty"
)

var (
	// Filled by the linker.
	fullVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

func main() {
	os.Exit(realMain("timeit", os.Args[1:], os.Stderr, nil))
}

type Cfg struct {
	showVersion    bool
	checkVersion   bool
	noColor        bool
	tickerDuration time.Duration
}

// Run the command specified as the first element of `args` and wait for it to terminate.
// If channel `started`  is not nil, send a message on it when the command has
// successfully started. Write our output to `out`, while the command output goes to
// stdout and stderr as usual. Use  `progname` in the command-line parsing help messages.
// Return the exit code to pass to os.Exit().
func realMain(progname string, args []string, out io.Writer, started chan<- (struct{})) int {
	flag.CommandLine.SetOutput(out)
	flagSet := flag.NewFlagSet(progname, flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(out, "%s measures the time of command execution\n\n", progname)
		fmt.Fprintf(out, "Usage: %s command\n\n", progname)
		fmt.Fprintf(out, "Options:\n")
		flagSet.PrintDefaults()
	}

	var cfg Cfg
	flagSet.BoolVar(&cfg.showVersion, "version", false, "show version")
	flagSet.BoolVar(&cfg.checkVersion, "check-version", false,
		"check online if new version is available")
	flagSet.BoolVar(&cfg.noColor, "no-color", false,
		"disable color output")
	flagSet.DurationVar(&cfg.tickerDuration, "ticker", 0,
		"print a status line each <duration>")

	if flagSet.Parse(args) != nil {
		return 2
	}
	if cfg.showVersion {
		fmt.Fprintln(out, "timeit version", fullVersion)
		return 0
	}
	if cfg.checkVersion {
		humanURL := fmt.Sprintf("https://github.com/%s/%s", "marco-m", "timeit")
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
			fmt.Fprintln(out, "To upgrade visit", humanURL)
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

	file, ok := out.(*os.File)
	if !ok || !isatty.IsTerminal(file.Fd()) || cfg.noColor {
		color.NoColor = true
	}

	return run(flagSet.Args()[0], flagSet.Args()[1:], cfg, out, started)
}

func run(
	progname string,
	args []string,
	cfg Cfg,
	out io.Writer,
	started chan<- (struct{}),
) int {
	chroma := color.New(color.FgCyan, color.Bold)

	cmd := exec.Command(progname, args...)
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
		for {
			sig := <-signalCh
			chroma.Fprintf(out, "\ntimeit: ignoring received signal: %v\n", sig)
		}
	}()

	if started != nil {
		go func() {
			started <- struct{}{}
		}()
	}

	if cfg.tickerDuration != 0 {
		ticker := time.NewTicker(cfg.tickerDuration)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				chroma.Fprintf(out, "\ntimeit ticker: running for %s\n",
					time.Since(start).Round(time.Millisecond))
			}
		}()
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		chroma.Fprintln(out, "timeit: wait child:", waitErr)
	}
	chroma.Fprintf(out, `
timeit results:
real: %v
`,
		time.Since(start).Round(time.Millisecond))
	code := cmd.ProcessState.ExitCode()
	if code == -1 {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		// Follow the shell convention, https://en.wikipedia.org/wiki/Exit_status
		code = 128 + int(status.Signal())
	}
	return code
}
