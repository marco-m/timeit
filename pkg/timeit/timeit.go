package timeit

// This code is released under the MIT License
// Copyright (c) 2020-23 Marco Molteni and the timeit contributors.

import (
	"fmt"
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

type Status struct {
	msg     string
	elapsed time.Duration
	code    int
}

func Main() int {
	var cfg config
	kong.Parse(&cfg,
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
		if err := checkVersion(); err != nil {
			fmt.Fprintf(os.Stderr, "timeit: %s\n", err)
			return 1
		}
		return 0
	}

	if len(cfg.Command) == 0 {
		fmt.Fprintf(os.Stderr, "timeit: expected <command> ...\n")
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

	chroma := color.New(color.FgMagenta, color.Bold)
	out := func(format string, a ...any) {
		chroma.Fprintf(os.Stderr, format, a...)
	}

	status := run(cmd, args, cfg, out)

	out(`
timeit results:
    %s
    real: %s
`, status.msg, status.elapsed.Round(time.Millisecond))

	return status.code
}

func checkVersion() error {
	const owner = "marco-m"
	const repo = "timeit"
	humanURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	latestVersion, err := release.GitHubLatest(owner, repo)
	if err != nil {
		return fmt.Errorf("checkVersion: %s", err)
	}
	result, err := release.Compare(shortVersion, latestVersion)
	if err != nil {
		return fmt.Errorf("checkVersion: %s", err)
	}
	switch result {
	case 0:
		fmt.Printf("timeit: installed version %s is the same as the latest version %s\n",
			shortVersion, latestVersion)
	case -1:
		fmt.Printf("timeit: installed version %s is older than the latest version %s\n",
			shortVersion, latestVersion)
		fmt.Println("To upgrade visit", humanURL)
	case +1:
		fmt.Printf("timeit: (unexpected?) installed version %s is newer than the latest version %s\n",
			shortVersion, latestVersion)
	}
	return nil
}

// Run executable (name, args) and wait for it to terminate.
// Write our output to `out`, while the command output goes to stdout and stderr as usual.
// Return the Status of the terminated executable.
func run(name string, args []string, cfg config, out func(format string, a ...any)) Status {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTS := time.Now()
	if err := cmd.Start(); err != nil {
		return Status{
			msg:     fmt.Sprintf("starting command: %s", err),
			elapsed: time.Since(startTS),
			code:    1,
		}
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
			out("\ntimeit: ignoring received signal: %v\n", sig)
		}
	}()

	if cfg.TickerDuration != 0 {
		ticker := time.NewTicker(cfg.TickerDuration)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				out("\ntimeit ticker: running for %s\n",
					time.Since(startTS).Round(cfg.TickerDuration))
			}
		}()
	}

	waitErr := cmd.Wait()

	elapsed := time.Since(startTS)
	code := cmd.ProcessState.ExitCode()
	switch code {
	case 0: // Success.
		return Status{
			msg:     "command succeeded",
			elapsed: elapsed,
			code:    code,
		}
	case -1: // Process was terminated by a signal.
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return Status{
			msg:     fmt.Sprintf("command terminated abnormally: %s", waitErr),
			elapsed: elapsed,
			// Follow the shell convention, https://en.wikipedia.org/wiki/Exit_status
			code: 128 + int(status.Signal()),
		}
	default: // Failure.
		return Status{
			msg:     fmt.Sprintf("command failed: %s", waitErr),
			elapsed: elapsed,
			code:    code,
		}
	}
}
