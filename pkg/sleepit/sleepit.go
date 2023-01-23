package sleepit

// This code is released under the MIT License
// Copyright (c) 2020-2023 Marco Molteni and the timeit contributors.

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
)

type SleepError struct {
	msg  string
	code int
}

func (e *SleepError) Error() string {
	return fmt.Sprintf("msg: %s code: %d", e.msg, e.code)
}

var (
	CleanerDoneError      = &SleepError{msg: "cleaner done", code: 3}
	CleanerCancelledError = &SleepError{msg: "cleaner cancelled", code: 4}
)

type common struct {
	Sleep time.Duration `help:"Sleep duration" default:"5s"`
}

type config struct {
	Common  common        `embed:""`
	Default SigDefaultCmd `cmd:"" help:"Use default signal action: on reception of SIGINT terminate abruptly."`
	Handle  SigHandleCmd  `cmd:"" help:"Handle signals: on reception of SIGINT perform cleanup before exiting."`
}

type SigDefaultCmd struct{}

type SigHandleCmd struct {
	Cleanup   time.Duration `help:"Cleanup duration" default:"5s"`
	TermAfter int           `help:"Terminate immediately after N signals. Default is to terminate only when the cleanup phase has completed." default:"0"`
}

func Main() int {
	var cfg config
	kctx := kong.Parse(&cfg,
		kong.Name("sleepit"),
		kong.Description("The sleepit utility sleeps for the specified duration, optionally handling signals.\n\nWhen the line \"sleepit: ready\" is printed, it means that it is safe to send signals to it.\n"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: false,
			Summary: true,
		}))
	if err := kctx.Run(cfg.Common); err != nil {
		var se *SleepError
		if errors.As(err, &se) {
			fmt.Fprintln(os.Stderr, "sleepit:", se.msg)
			return se.code
		}
		fmt.Fprintln(os.Stderr, "unexpected:", err)
		return 100 // fixme
	}
	return 0
}

func (cmd *SigDefaultCmd) Run(ctx common) error {
	return supervisor(ctx.Sleep, 0, 0, nil)
}

func (cmd *SigHandleCmd) Run(ctx common) error {
	if cmd.TermAfter == 1 {
		return &SleepError{
			msg:  "handle: --term-after cannot be 1",
			code: 2,
		}
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt) // Ctrl-C -> SIGINT
	return supervisor(ctx.Sleep, cmd.Cleanup, cmd.TermAfter, sigCh)
}

func supervisor(
	sleep time.Duration,
	cleanup time.Duration,
	termAfter int,
	sigCh <-chan os.Signal,
) error {
	fmt.Printf("sleepit: ready\n")
	fmt.Printf("sleepit: PID=%d sleep=%v cleanup=%v\n",
		os.Getpid(), sleep, cleanup)

	cancelWork := make(chan struct{})
	workerDone := worker(cancelWork, sleep, "work")

	cancelCleaner := make(chan struct{})
	var cleanerDone <-chan struct{}

	sigCount := 0
	for {
		select {
		case sig := <-sigCh:
			sigCount++
			fmt.Printf("sleepit: got signal=%s count=%d\n", sig, sigCount)
			if sigCount == 1 {
				// since `cancelWork` is unbuffered, sending will be synchronous:
				// we are ensured that the worker has terminated before starting cleanup.
				// This is important in some real-life situations.
				cancelWork <- struct{}{}
				cleanerDone = worker(cancelCleaner, cleanup, "cleanup")
			}
			if sigCount == termAfter {
				cancelCleaner <- struct{}{}
				return CleanerCancelledError
			}
		case <-workerDone:
			return nil
		case <-cleanerDone:
			return CleanerDoneError
		}
	}
}

// Start a worker goroutine and return immediately a `workerDone` channel.
// The goroutine will prepend its prints with the prefix `name`.
// The goroutine will simulate some work and will terminate when one of the following
// conditions happens:
//  1. When `howlong` is elapsed. This case will be signaled on the `workerDone` channel.
//  2. When something happens on channel `canceled`. Note that this simulates real-life,
//     so cancellation is not instantaneous: if the caller wants a synchronous cancel,
//     it should send a message; if instead it wants an asynchronous cancel, it should
//     close the channel.
func worker(
	canceled <-chan struct{},
	howlong time.Duration,
	name string,
) <-chan struct{} {
	workerDone := make(chan struct{})
	deadline := time.Now().Add(howlong)
	sleep := 100 * time.Millisecond
	if howlong < sleep {
		sleep = howlong / 2
	}

	go func() {
		fmt.Printf("sleepit: %s started\n", name)
		for {
			select {
			case <-canceled:
				fmt.Printf("sleepit: %s canceled\n", name)
				return
			default:
				if doSomeWork(deadline, sleep) {
					fmt.Printf("sleepit: %s done\n", name) // <== NOTE THIS LINE
					workerDone <- struct{}{}
					return
				}
			}
		}
	}()
	return workerDone
}

// Do some work for sleep duration and then return, so that the caller can decide whether
// to continue or not.
// Return true when all work is done.
func doSomeWork(deadline time.Time, sleep time.Duration) bool {
	if time.Now().After(deadline) {
		return true
	}
	time.Sleep(sleep)
	return false
}
