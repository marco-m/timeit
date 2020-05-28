// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

const usage = "usage: sleepit <duration>"

func main() {
	os.Exit(run(os.Args, os.Stderr))
}

func run(args []string, out io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(out, usage)
		return 2
	}

	duration, err := time.ParseDuration(args[1])
	if err != nil {
		fmt.Fprintln(out, err)
		fmt.Fprintln(out, usage)
		return 2
	}

	time.Sleep(duration)
	return 0
}
