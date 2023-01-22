package timeit

// This code is released under the MIT License
// Copyright (c) 2020-23 Marco Molteni and the timeit contributors.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/marco-m/taschino/pkg/release"
	"github.com/mattn/go-isatty"
)

var (
	// Filled by the linker.
	fullVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

type config struct {
	Version        bool          `help:"Display version and exit."`
	CheckVersion   bool          `help:"Check online if new version is available and exit."`
	NoColor        bool          `help:"Disable color output."`
	TickerDuration time.Duration `name:"ticker" placeholder:"DURATION" help:"Print a status line each DURATION."`

	// Command must be optional to support --version
	Command []string `arg:"" optional:"" passthrough:"" help:"Command to time."`
}

func Main() int {
	var cfg config
	kctx := kong.Parse(&cfg,
		kong.Name("timeit"),
		kong.Description("The timeit utility measures the time of command execution."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: false,
			Summary: true,
		}))

	if cfg.Version {
		fmt.Println("timeit:")
		fmt.Printf("  version: %s\n  home:    https://github.com/marco-m/timeit\n",
			fullVersion)
		return 0
	}

	if cfg.CheckVersion {
		return checkVersion()
	}

	if len(cfg.Command) == 0 {
		kctx.Errorf("expected <command> ...")
		return 1
	}

	if !isatty.IsTerminal(os.Stderr.Fd()) || cfg.NoColor {
		color.NoColor = true
	}

	cmd := cfg.Command[0]
	var args []string
	if len(cfg.Command) > 1 {
		args = cfg.Command[1:]
	}
	return run(cmd, args, cfg, os.Stderr)
}

func checkVersion() int {
	const owner = "marco-m"
	const repo = "timeit"
	humanURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	latestVersion, err := release.GitHubLatest(owner, repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	result, err := release.Compare(shortVersion, latestVersion)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	switch result {
	case 0:
		fmt.Fprintf(os.Stderr,
			"timeit installed version %s is the same as the latest version %s\n",
			shortVersion, latestVersion)
	case -1:
		fmt.Fprintf(os.Stderr,
			"timeit installed version %s is older than the latest version %s\n",
			shortVersion, latestVersion)
		fmt.Fprintln(os.Stderr, "To upgrade visit", humanURL)
	case +1:
		fmt.Fprintf(os.Stderr,
			"timeit (unexpected?) installed version %s is newer than the latest version %s\n",
			shortVersion, latestVersion)
	}
	return 0
}

// Run progname and wait for it to terminate.
// Write our output to `out`, while the command output goes to stdout and stderr as usual.
// Return the exit code to pass to os.Exit().
func run(progname string, args []string, cfg config, out io.Writer) int {
	chroma := color.New(color.FgMagenta, color.Bold)

	cmd := exec.Command(progname, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		chroma.Fprintln(out, "timeit: starting child:", err)
		return 1
	}

	// We are in the parent, after having started the child.
	// Ignoring SIGINT as the original /usr/bin/time does with
	// signal.Ignore(os.Interrupt) has subtle side effects with the tests.
	// Thus, we do the equivalent with a do-nothing signal handler.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			sig := <-signalCh
			chroma.Fprintf(out, "\ntimeit: ignoring received signal: %v\n", sig)
		}
	}()

	if cfg.TickerDuration != 0 {
		ticker := time.NewTicker(cfg.TickerDuration)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				chroma.Fprintf(out, "\ntimeit ticker: running for %s\n",
					time.Since(start).Round(cfg.TickerDuration))
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		chroma.Fprintln(out, "timeit: wait child:", err)
	}
	chroma.Fprintf(out, "\ntimeit results:\nreal: %v\n",
		time.Since(start).Round(time.Millisecond))

	code := cmd.ProcessState.ExitCode()
	if code == -1 {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		// Follow the shell convention, https://en.wikipedia.org/wiki/Exit_status
		code = 128 + int(status.Signal())
	}
	return code
}
