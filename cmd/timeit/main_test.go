// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
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
			127,
			"timeit: exec: \"non-existing\": executable file not found in $PATH",
			false,
		},
		{
			"non existing command is an error, absolute path",
			[]string{"/non-existing"},
			127,
			"timeit: exec: \"/non-existing\": stat /non-existing: no such file or directory",
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
			gotCode := run("timeit", tc.args, &gotOut)
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
