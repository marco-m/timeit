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
	handleSleep := handleCmd.Duration("sleep", 5*time.Second, "Sleep duration")
	handleCleanup := handleCmd.Duration("cleanup", 5*time.Second, "Cleanup duration")

	handleTermAfter := handleCmd.Int("term-after", 0,
		"Terminate immediately after `N` signals.\n"+
			"Default is to terminate only when the cleanup phase has completed.")

	defaultCmd := flag.NewFlagSet("default", flag.ExitOnError)
	defaultSleep := defaultCmd.Duration("sleep", 5*time.Second, "Sleep duration")

	switch args[0] {
	case "handle":
		handleCmd.Parse(args[1:])
		if *handleTermAfter == 1 {
			fmt.Fprintf(out, "handle: term-after cannot be 1\n")
			return 2
		}
		if len(handleCmd.Args()) > 0 {
			fmt.Fprintf(out, "handle: unexpected arguments: %v\n", handleCmd.Args())
			return 2
		}
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt) // Ctrl-C -> SIGINT
		return supervisor(out, *handleSleep, *handleCleanup, *handleTermAfter, sigCh)

	case "default":
		defaultCmd.Parse(args[1:])
		if len(defaultCmd.Args()) > 0 {
			fmt.Fprintf(out, "default: unexpected arguments: %v\n", defaultCmd.Args())
			return 2
		}
		return supervisor(out, *defaultSleep, 0, 0, nil)

	default:
		fmt.Fprintln(out, usage)
		return 2
	}
}

func supervisor(
	out io.Writer,
	sleep time.Duration,
	cleanup time.Duration,
	termAfter int,
	sigCh <-chan os.Signal,
) int {
	fmt.Fprintf(out, "sleepit: ready\n")
	fmt.Fprintf(out, "sleepit: PID=%d sleep=%v cleanup=%v\n",
		os.Getpid(), sleep, cleanup)

	cancelWork := make(chan struct{})
	workerDone := worker(cancelWork, sleep)

	cancelCleaner := make(chan struct{})
	var cleanerDone <-chan struct{}

	sigCount := 0
	for {
		select {
		case sig := <-sigCh:
			sigCount++
			fmt.Fprintf(out, "sleepit: got signal=%s count=%d\n", sig, sigCount)
			if sigCount == 1 {
				// since `cancelWork` is unbuffered, sending will be synchronous:
				// we are ensured that the worker has terminated before starting cleanup.
				// This is important in some real-life situations.
				cancelWork <- struct{}{}
				cleanerDone = cleaner(cancelCleaner, cleanup)
			}
			if sigCount == termAfter {
				cancelCleaner <- struct{}{}
				return 4
			}
		case <-workerDone:
			return 0
		case <-cleanerDone:
			return 3
		}
	}
}

// Start a worker goroutine and return immediately a `workerDone` channel.
// The goroutine will simulate some work and will terminate when one of the following
// conditions happens:
// 1. When `howlong` is elapsed. This case will be signaled on the `workerDone` channel.
// 2. When something happens on channel `canceled`. Note that this simulates real-life,
//    so cancellation is not instantaneous: if the caller wants a synchronous cancel,
//    it should send a message; if instead it wants an asynchronous cancel, it should
//    close the channel.
func worker(canceled <-chan struct{}, howlong time.Duration) <-chan struct{} {
	workerDone := make(chan struct{})
	deadline := time.Now().Add(howlong)
	go func() {
		fmt.Printf("sleepit: work started\n")
		for {
			select {
			case <-canceled:
				fmt.Printf("sleepit: work canceled\n")
				return
			default:
				if doSomeWork(deadline) {
					fmt.Printf("sleepit: work done\n")
					workerDone <- struct{}{}
					return
				}
			}
		}
	}()
	return workerDone
}

// Do some work and then return, so that the caller can decide wether to continue or not.
// Return true when all work is done.
func doSomeWork(deadline time.Time) bool {
	if time.Now().After(deadline) {
		return true
	}
	timeout := 100 * time.Millisecond
	time.Sleep(timeout)
	return false
}

// Start a cleaner goroutine and return immediately a `cleanerDone` channel.
// The goroutine will simulate cleaning up for `cleanup` duration and will signal on
// channel `cleanerDone` when it has terminated.
func cleaner(canceled <-chan struct{}, howlong time.Duration) <-chan struct{} {
	cleanerDone := make(chan struct{})
	go func() {
		fmt.Printf("sleepit: cleanup started, please wait\n")
		select {
		case <-canceled:
			fmt.Printf("sleepit: cleanup canceled\n")
		case <-time.After(howlong):
			fmt.Printf("sleepit: cleanup done\n") // <== NOTE THIS LINE
			cleanerDone <- struct{}{}
		}
	}()
	return cleanerDone
}
