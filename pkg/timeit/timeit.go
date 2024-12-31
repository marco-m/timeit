package timeit

// This code is released under the MIT License
// Copyright (c) 2020-23 Marco Molteni and the timeit contributors.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/marco-m/taschino/pkg/release"
	"github.com/mattn/go-isatty"
)

var (
	// Filled by the linker.
	longVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

type config struct {
	Version        bool          `help:"Display version and exit."`
	CheckVersion   bool          `help:"Check online if new version is available and exit."`
	NoColor        bool          `help:"Disable color output."`
	TickerDuration time.Duration `name:"ticker" placeholder:"DURATION" help:"Print a status line each DURATION."`
	Observe        string        `placeholder:"FORMAT" help:"observe the output according to FORMAT and print a summary on each ticker. Supported formats: pytest."`

	// Command must be optional to support --version
	Command []string `arg:"" optional:"" passthrough:"" help:"Command to time."`
}

type printFn func(format string, a ...any)

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
			longVersion)
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

	if cfg.Observe != "" {
		if cfg.Observe != "pytest" {
			fmt.Fprintf(os.Stderr,
				"timeit: unknown --observe=%s; must be pytest\n", cfg.Observe)
			return 1
		}
	}
	if cfg.Observe != "" && cfg.TickerDuration == 0 {
		fmt.Fprintf(os.Stderr,
			"timeit: --observe requires --ticker\n")
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

	return run(cmd, args, cfg, out)
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

type event struct {
	name     string
	started  time.Time
	finished time.Time
}

type records struct {
	mu sync.Mutex
	// Event name -> event data
	flying map[string]event
	// Event name -> event data
	landed map[string]event
}

func newRecords() *records {
	return &records{
		flying: make(map[string]event, 100),
		landed: make(map[string]event, 100),
	}
}

// Run executable (name, args) and wait for it to terminate.
// Write our output to `out`, while the command output goes to stdout and stderr as usual.
// Return the status code of the terminated executable.
func run(name string, args []string, cfg config, out printFn) int {
	dur100 := cfg.TickerDuration / 100
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		var elapsed time.Duration = 0
		out(results(fmt.Sprintf("getting pipe for command stdout: %s", err), elapsed, dur100, nil))
		return 1
	}

	t0 := time.Now()
	if err := cmd.Start(); err != nil {
		elapsed := time.Since(t0)
		out(results(fmt.Sprintf("starting command: %s", err), elapsed, dur100, nil))
		return 1
	}

	//
	// Here we are in the parent, after having started the child.
	//

	records := newRecords()
	setupProcessOutput(cfg.Observe, records, stdout, out)

	setupSignalHandling(out)

	cancelTicker := setupPeriodicTicker(t0, cfg.TickerDuration, cfg.Observe != "", records, out)

	// When using pipes, cmd.Wait() must be called _after_ the pipe is drained.
	// See https://pkg.go.dev/os/exec#Cmd.StdoutPipe
	// <-done FIXME
	waitErr := cmd.Wait()
	elapsed := time.Since(t0)
	cancelTicker()

	msg, code := extractStatus(cmd.ProcessState, waitErr)
	out("%s", results(msg, elapsed, dur100, records))
	return code
}

func results(msg string, elapsed time.Duration, precision time.Duration, records *records) string {
	var bld strings.Builder

	fmt.Fprintf(&bld, `
timeit results:
    %s
    real: %s
`, msg, elapsed.Round(time.Millisecond))

	if records != nil {
		fmt.Fprintf(&bld, "    flights by duration:\n")
		tw := tabwriter.NewWriter(&bld, 5, 0, 2, ' ', 0)
		landed := make([]event, 0, len(records.landed))

		// From map to slice, so that we can sort by duration.
		for _, evt := range records.landed {
			landed = append(landed, evt)
		}

		sort.Slice(landed, func(i, j int) bool {
			elapsedI := landed[i].finished.Sub(landed[i].started)
			elapsedJ := landed[j].finished.Sub(landed[j].started)
			return elapsedI > elapsedJ
		})
		for i, evt := range landed {
			elapsed := evt.finished.Sub(evt.started).Truncate(precision)
			fmt.Fprintf(tw, "    %4d\t%s\t%8v\n", i+1, evt.name, elapsed)
		}
		tw.Flush()
	}

	return bld.String()
}

// FIXME done channel !!! ALL methods...
// FIXME add also documentation...
func setupProcessOutput(observe string, events *records, stdout io.Reader, out printFn) {
	switch observe {
	case "pytest":
		go observePytest(events, stdout, out)

	// Simple stdout copier if --observe flag is missing or unknown.
	default:
		go func() {
			if _, err := io.Copy(os.Stdout, stdout); err != nil {
				// FIXME Report to the errors channel and be printed at the end.
				out("timeit: copying stdout: %s\n", err)
			}
		}()
	}
}

// We are in the parent, after having started the child.
// Ignoring SIGINT as the original /usr/bin/time does with
// signal.Ignore(os.Interrupt) has subtle side effects with the tests.
// Thus, we do the equivalent with a do-nothing signal handler.
func setupSignalHandling(out printFn) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		count := 0
		for {
			sig := <-signalCh
			count++
			out("timeit: got signal name=%s count=%d disposition=ignore\n", sig, count)
		}
	}()
}

func setupPeriodicTicker(t0 time.Time, dur time.Duration, summarize bool, records *records, out printFn) func() {
	if dur == 0 {
		return func() {}
	}

	done := make(chan struct{})
	ticker := time.NewTicker(dur)
	dur100 := dur / 100
	var bld strings.Builder
	tw := tabwriter.NewWriter(&bld, 5, 0, 2, ' ', 0)
	flying := make([]event, 0, 100)

	go func() {
		for {
			select {

			case <-done:
				ticker.Stop()
				return

			case now := <-ticker.C:
				fmt.Fprintf(&bld, "\ntimeit ticker: running for %s\n",
					now.Sub(t0).Truncate(dur))
				if summarize {
					fmt.Fprintf(&bld, "in-flight:\n")

					records.mu.Lock()
					// From map to slice, so that we can sort by duration.
					for _, evt := range records.flying {
						flying = append(flying, evt)
					}
					records.mu.Unlock()

					sort.Slice(flying, func(i, j int) bool {
						return flying[i].started.Before(flying[j].started)
					})
					for i, evt := range flying {
						elapsed := now.Sub(evt.started).Truncate(dur100)
						fmt.Fprintf(tw, "    %4d\t%s\t%6v\n", i+1, evt.name, elapsed)
					}
					tw.Flush()
				}
				out("%s\n", bld.String())
				bld.Reset()
				// reset, keep allocated memory
				flying = flying[:0]
			}
		}
	}()

	return func() {
		done <- struct{}{}
	}
}

func extractStatus(procState *os.ProcessState, waitErr error) (string, int) {
	code := procState.ExitCode()
	switch code {
	case 0: // Success.
		return "command succeeded", code
	case -1: // Process was terminated by a signal.
		// Follow the shell convention, https://en.wikipedia.org/wiki/Exit_status
		status := procState.Sys().(syscall.WaitStatus)
		code := 128 + int(status.Signal())
		return fmt.Sprintf("command terminated abnormally: %s", waitErr), code
	default: // Failure.
		return fmt.Sprintf("command failed: %s", waitErr), code
	}
}
