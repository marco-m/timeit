// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	SLEEPIT, _ = filepath.Abs("../../bin/sleepit")
	TIMEIT, _  = filepath.Abs("../../bin/timeit")
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
			[]string{SLEEPIT, "10ms", "0s"},
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
			gotCode := realMain("timeit", tc.args, &gotOut, nil)
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

func TestSignalSentToProcessGroup(t *testing.T) {
	var out bytes.Buffer
	sut := execCommand(TIMEIT, SLEEPIT, "2s", "10ms")
	sut.Stdout = &out
	sut.Stderr = &out

	// Create a new process group, by setting the process group ID of the child to the
	// child PID.
	// By default, the child would inherit the process group of the parent, but we want
	// to avoid this, to protect the parent (the test process) from the signal that this
	// test will send. More info in the comments below for syscall.Kill().
	sut.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}

	if err := sut.Start(); err != nil {
		t.Fatalf("starting the timeit process: %v", err)
	}

	// After the child is started, we want to avoid a race condition where we send it a
	// signal before it had time to setup its own signal handlers.
	// Sleeping is way too flaky, instead we parse the child output until we get a line
	// that we know is printed after the signal handlers are installed...
	ready := false
	timeout := time.Duration(time.Second)
	start := time.Now()
	for time.Since(start) < timeout {
		if strings.Contains(out.String(), "sleepit: ready") {
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("sleepit not ready after %v", timeout)
	}

	// When we have a running program in a shell and type CTRL-C, the tty driver will
	// send a SIGINT signal to all the processes in the foreground process group
	// (see https://en.wikipedia.org/wiki/Process_group).
	//
	// Here we want to emulate this behavior: send SIGINT to the process group of the
	// test executable. Although Go for some reasons doesn't wrap the killpg(2) system
	// call, what works is using syscall.Kill(-PID, SIGINT), where the negative PID means
	// the corresponding process group. Note that this negative PID works only as long
	// as the caller of the kill(2) system call has a different PID, which is the case
	// for this test.
	if err := syscall.Kill(-sut.Process.Pid, syscall.SIGINT); err != nil {
		t.Fatalf("sending INT signal to the process group: %v", err)
	}

	err := sut.Wait()

	var wantErr *exec.ExitError
	const wantExitStatus = 3 // sleepit returns 3 if it receives SIGINT
	if errors.As(err, &wantErr) {
		if wantErr.ExitCode() != wantExitStatus {
			t.Errorf("waiting for the timeit process: got exit status %v; want %d",
				wantErr.ExitCode(), wantExitStatus)
			t.Errorf("exited normally (that is: not terminated by a signal): %v",
				wantErr.Exited())
			t.Errorf("Process state: %q", wantErr.String())
		}
	} else {
		t.Errorf("waiting for the timeit process: got %v (%T); want (%T)", err, err, wantErr)
	}

	wantMsg := "sleepit: cleanup done"
	if !strings.Contains(out.String(), wantMsg) {
		t.Errorf("output: %q does not contain %q", out.String(), wantMsg)
	}
}

func TestShowVersion(t *testing.T) {
	var gotOut bytes.Buffer
	if gotCode := realMain("timeit", []string{"-version"}, &gotOut, nil); gotCode != 0 {
		t.Errorf("\ngotCode: %v; want: 0", gotCode)
	}
	wantPrefix := "timeit version "
	if !strings.HasPrefix(gotOut.String(), wantPrefix) {
		t.Errorf("\ngotOut: %s;wantPrefix: %s", gotOut.String(), wantPrefix)
	}
}
