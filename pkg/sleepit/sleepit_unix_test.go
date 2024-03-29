//go:build !windows

package sleepit_test

import (
	"bytes"
	"errors"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestSignalSentToProcessGroup(t *testing.T) {
	testCases := map[string]struct {
		args     []string
		sendSigs int
		wantOut  []string
	}{
		"send 1 signal": {
			args:     []string{"handle", "--sleep=10ms", "--cleanup=10ms"},
			sendSigs: 1,
			wantOut: []string{
				"sleepit: ready\n",
				"sleepit: work started\n",
				"sleepit: got signal=interrupt count=1\n",
				"sleepit: work canceled\n",
				"sleepit: cleanup started\n",
				"sleepit: cleanup done\n"},
		},
		"send 2 signals": {
			args:     []string{"handle", "--sleep=10ms", "--cleanup=10ms"},
			sendSigs: 2,
			wantOut: []string{
				"sleepit: ready\n",
				"sleepit: work started\n",
				"sleepit: got signal=interrupt count=1\n",
				"sleepit: got signal=interrupt count=2\n",
				"sleepit: work canceled\n",
				"sleepit: cleanup started\n",
				"sleepit: cleanup done\n"},
		},
	}

	tmpDir := t.TempDir()

	sleepit := path.Join(tmpDir, "sleepit")
	cmd := exec.Command("go", "build", "-o", sleepit, "../../cmd/sleepit")
	assert.NilError(t, cmd.Run())

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			sut := exec.Command(sleepit, tc.args...)
			sut.Stdout = &out
			sut.Stderr = &out
			// Create a new process group by setting the process group ID of the child
			// to the child PID.
			// By default, the child would inherit the process group of the parent, but
			// we want to avoid this, to protect the parent (the test process) from the
			// signal that this test will send. More info in the comments below for
			// syscall.Kill().
			sut.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}

			if err := sut.Start(); err != nil {
				t.Fatalf("starting the SUT process: %v", err)
			}

			// After the child is started, we want to avoid a race condition where we send
			// it a signal before it had time to set up its own signal handlers. Sleeping
			// is way too flaky, instead we parse the child output until we get a line
			// that we know is printed after the signal handlers are installed...
			ready := false
			timeout := time.Duration(time.Second)
			start := time.Now()
			for time.Since(start) < timeout {
				if strings.Contains(out.String(), "sleepit: ready\n") {
					ready = true
					break
				}
				time.Sleep(time.Millisecond)
			}
			if !ready {
				t.Fatalf(
					"sleepit not ready after %v\n"+
						"additional information:\n"+
						"  out:\n"+
						"%s",
					timeout,
					out.String())
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
			for i := 1; i <= tc.sendSigs; i++ {
				if err := syscall.Kill(-sut.Process.Pid, syscall.SIGINT); err != nil {
					t.Fatalf("sending INT signal to the process group: %v", err)
				}
				time.Sleep(1 * time.Millisecond)
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

			gotLines := strings.SplitAfter(out.String(), "\n")
			notFound := notContained(gotLines, tc.wantOut)
			if len(notFound) > 0 {
				t.Errorf("some lines not found in output:\nnot found: %q\noutput: %q",
					notFound, gotLines)
			}
		})
	}
}

// Return a list of lines from `wantLines` that are not contained in `gotLines`.
// FIXME this does not enforce ordering. We might want to support both.
func notContained(gotLines []string, wantLines []string) []string {
	notFound := []string{}

	for _, wantLine := range wantLines {
		found := false
		for _, gotLine := range gotLines {
			if strings.Contains(gotLine, wantLine) {
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, wantLine)
		}
	}

	return notFound
}
