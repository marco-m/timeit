// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

const usage = "usage: sleepit <sleep-duration> <cleanup-duration>"

func main() {
	os.Exit(run(os.Args, os.Stderr))
}

func run(args []string, out io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(out, usage)
		return 2
	}

	sleepDuration, err := time.ParseDuration(args[1])
	if err != nil {
		fmt.Fprintln(out, "parsing sleep-duration:", err)
		fmt.Fprintln(out, usage)
		return 2
	}

	cleanupDuration, err := time.ParseDuration(args[2])
	if err != nil {
		fmt.Fprintln(out, "parsing cleanup-duration:", err)
		fmt.Fprintln(out, usage)
		return 2
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt) // Ctrl-C -> SIGINT
	fmt.Printf("sleepit: ready:\n")
	fmt.Printf("sleepit: do one of the following:\n")
	fmt.Printf("- From this shell:    type: CTRL-C\n")
	fmt.Printf("- From another shell: type: kill -INT %d\n", os.Getpid())

	fmt.Printf("sleepit: working...\n")
	select {
	case <-time.After(sleepDuration):
		fmt.Printf("sleepit: work done\n")
	case sig := <-sigCh:
		fmt.Printf("sleepit: got signal: %v\n", sig)
		fmt.Printf("sleepit: cleaning up, please wait\n")
		time.Sleep(cleanupDuration)
		fmt.Printf("sleepit: cleanup done\n") // <== NOTE THIS LINE
		return 3
	}

	return 0
}
