// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	SLEEPIT, _ = filepath.Abs("../../bin/sleepit")
)

func TestRun(t *testing.T) {
	testCases := []struct {
		description     string
		args            []string
		wantCode        int
		wantErrLine     string
		wantResultsLine bool
	}{
		{
			"missing command is an error",
			[]string{},
			2,
			"timeit: expected a command to time",
			false,
		},
		{
			"non existing command is an error, relative path",
			[]string{"non-existing"},
			1,
			"timeit: start child: exec: \"non-existing\": executable file not found in $PATH",
			false,
		},
		{
			"non existing command is an error, absolute path",
			[]string{"/non-existing"},
			1,
			"timeit: start child: fork/exec /non-existing: no such file or directory",
			false,
		},
		{
			"child status 0 is forwarded",
			[]string{SLEEPIT, "10ms"},
			0,
			"",
			true,
		},
		{
			"child status 2 is forwarded",
			[]string{SLEEPIT, "x"},
			2,
			"timeit: wait child: exit status 2",
			true,
		},
		{
			"/usr/bin/false child status 1 is forwarded",
			[]string{"false"},
			1,
			"timeit: wait child: exit status 1",
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var gotOut bytes.Buffer
			gotCode := run("timeit", tc.args, &gotOut, nil)
			if gotCode != tc.wantCode {
				t.Errorf("\ncode: got: %d; want: %d", gotCode, tc.wantCode)
			}
			gotErrLine, containsResults := parseOutput(gotOut.String())
			if tc.wantResultsLine != containsResults {
				t.Errorf("\nwantResults: %v; containsResults: %v\noutput: %q",
					tc.wantResultsLine, containsResults, gotOut.String())
			}
			if gotErrLine != tc.wantErrLine {
				t.Errorf("\nerrline: got: %q; want: %q", gotErrLine, tc.wantErrLine)
			}
		})
	}
}

// This parsing is lenient but I think it is good enough.
// Expects output to be as follows, with each line optional:
//   ...
//   timeit: error message
//   timeit results:
//   ...
//
func parseOutput(out string) (timeitErrLine string, containsTimeitResults bool) {
	var errLine string
	var containsResults bool
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "timeit:") {
			errLine = line
		}
		if strings.HasPrefix(line, "timeit results:") {
			containsResults = true
			break
		}
	}
	return errLine, containsResults
}

func TestReturnCorrectExitCodeIfChildTerminatedBySignal(t *testing.T) {
	// We use the TestHelperProcess pattern (see below) and ask the child to send a
	// signal to itself.
	execCommand = helperCommand
	defer func() {
		execCommand = exec.Command
	}()

	var gotOut bytes.Buffer
	gotCode := run("timeit", []string{"signal"}, &gotOut, nil)
	wantCode := 128 + int(syscall.SIGINT)
	if gotCode != wantCode {
		t.Fatalf("\ncode: got: %d; want: %d\nout: %q", gotCode, wantCode, gotOut.String())
	}
}

// The TestHelperProcess pattern, part 1.
// See TestHelperProcess() for more info.
func helperCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

// The TestHelperProcess pattern, part 2.
// See:
// - helperCommand()
// - https://golang.org/src/os/exec/exec_test.go
// - https://npf.io/2015/06/testing-exec-command
//
// TestHelperProcess isn't a real test. It's used as a helper process.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// os.Args has been built by helperCommand() and is something like:
	// [/path/to/test-executable, -test.run=TestHelperProcess, --, cmd, args]
	// where cmd is the command and args is the list of arguments passed to
	// exec.Command() in the SUT.
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "TestHelperProcess: missing command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "signal":
		// We are now the child process of timeit.
		// Send ourselves a signal and thus terminate.
		self, _ := os.FindProcess(os.Getpid())
		if err := self.Signal(os.Interrupt); err != nil {
			fmt.Fprintf(os.Stderr, "TestHelperProcess: sending signal: %v\n", err)
			os.Exit(3)
		}
	default:
		fmt.Fprintf(os.Stderr, "TestHelperProcess: unknown cmd: %q; args: %q\n",
			cmd, args)
		os.Exit(4)
	}
}

func TestIgnoreSignals(t *testing.T) {
	// Normally sending SIGINT to the process running the test would cause process
	// termination and thus `go test` would report a failure. Since the SUT ignores
	// this signal, the test should pass.
	//
	started := make(chan struct{})

	// We use a goroutine because we want to send the signal while the SUT is running.
	go func() {
		select {
		case <-started:
			self, _ := os.FindProcess(os.Getpid())
			if err := self.Signal(os.Interrupt); err != nil {
				t.Errorf("sending signal: %v", err)
			}
		case <-time.After(time.Second):
			t.Errorf("timer expired and child not started")
		}
	}()

	var gotOut bytes.Buffer
	// FIXME can I have a better synchronization than a sleep ? :-(
	if gotCode := run("timeit", []string{SLEEPIT, "200ms"}, &gotOut, started); gotCode != 0 {
		t.Fatalf("\ngotCode: %v; want: 0", gotCode)
	}
}

func TestShowVersion(t *testing.T) {
	var gotOut bytes.Buffer
	if gotCode := run("timeit", []string{"-version"}, &gotOut, nil); gotCode != 0 {
		t.Errorf("\ngotCode: %v; want: 0", gotCode)
	}
	wantPrefix := "timeit version "
	if !strings.HasPrefix(gotOut.String(), wantPrefix) {
		t.Errorf("\ngotOut: %s;wantPrefix: %s", gotOut.String(), wantPrefix)
	}
}
