// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

const usage = `sleepit: sleep for the specified duration, optionally handling signals
When the line "sleepit: ready" is printed, it means that it is safe to send signals to it

Usage: sleepit <command> [<args>]

Commands

  handle      Handle signals: on reception of SIGINT perform cleanup before exiting
  default     Use default action: on reception of SIGINT terminate abruptly
`

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, out io.Writer) int {
	flag.CommandLine.SetOutput(out)
	if len(args) < 1 {
		fmt.Fprintln(out, usage)
		return 2
	}

	handleCmd := flag.NewFlagSet("handle", flag.ExitOnError)
	handleSleep := handleCmd.Duration("sleep", 5*time.Second, "sleep duration")
	handleCleanup := handleCmd.Duration("cleanup", 5*time.Second, "cleanup duration")

	defaultCmd := flag.NewFlagSet("default", flag.ExitOnError)
	defaultSleep := defaultCmd.Duration("sleep", 5*time.Second, "sleep duration")

	switch args[0] {
	case "handle":
		handleCmd.Parse(args[1:])
		if len(handleCmd.Args()) > 0 {
			fmt.Fprintf(out, "handle: unexpected arguments: %v\n", handleCmd.Args())
			return 2
		}
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt) // Ctrl-C -> SIGINT
		return doWork(out, *handleSleep, *handleCleanup, sigCh)

	case "default":
		defaultCmd.Parse(args[1:])
		if len(defaultCmd.Args()) > 0 {
			fmt.Fprintf(out, "default: unexpected arguments: %v\n", defaultCmd.Args())
			return 2
		}
		return doWork(out, *defaultSleep, 0, nil)

	default:
		fmt.Fprintln(out, usage)
		return 2
	}
}

func doWork(
	out io.Writer,
	sleep time.Duration,
	cleanup time.Duration,
	sigCh <-chan os.Signal,
) int {
	fmt.Fprintf(out, "sleepit: ready\n")
	fmt.Fprintf(out, "sleepit: PID=%d sleep=%v cleanup=%v\n",
		os.Getpid(), sleep, cleanup)

	select {
	case <-time.After(sleep):
		fmt.Fprintf(out, "sleepit: work done\n")
		return 0
	case sig := <-sigCh:
		fmt.Fprintf(out, "sleepit: got signal: %v\n", sig)
		fmt.Fprintf(out, "sleepit: cleaning up, please wait\n")
		time.Sleep(cleanup)
		fmt.Fprintf(out, "sleepit: cleanup done\n") // <== NOTE THIS LINE
		return 3
	}
}
