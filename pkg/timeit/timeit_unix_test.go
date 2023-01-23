//go:build !windows
// +build !windows

package timeit_test

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
	tmpDir := t.TempDir()

	timeit := path.Join(tmpDir, "timeit")
	cmd1 := exec.Command("go", "build", "-o", timeit, "../../cmd/timeit")
	assert.NilError(t, cmd1.Run())

	sleepit := path.Join(tmpDir, "sleepit")
	cmd2 := exec.Command("go", "build", "-o", sleepit, "../../cmd/sleepit")
	assert.NilError(t, cmd2.Run())

	var out bytes.Buffer
	sut := exec.Command(timeit, "--", sleepit, "handle", "--sleep=2s", "--cleanup=10ms")
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
		if strings.Contains(out.String(), "sleepit: ready\n") {
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
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
