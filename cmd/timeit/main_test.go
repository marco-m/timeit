// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

var (
	SLEEPIT, _ = filepath.Abs("../../bin/sleepit")
	PROGNAME   = "timeit"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		description string
		args        []string
		wantCode    int
		wantOut     string
	}{
		{
			"missing command is an error",
			[]string{},
			2,
			"timeit: expected a command to time\n",
		},
		{
			"non existing command is an error, relative path",
			[]string{"non-existing"},
			127,
			"timeit: exec: \"non-existing\": executable file not found in $PATH\n",
		},
		{
			"non existing command is an error, absolute path",
			[]string{"/non-existing"},
			127,
			"timeit: exec: \"/non-existing\": stat /non-existing: no such file or directory\n",
		},
		{
			"child status 0 is forwarded",
			[]string{SLEEPIT, "10ms"},
			0,
			"",
		},
		{
			"child status 2 is forwarded",
			[]string{SLEEPIT, "x"},
			2,
			"timeit: wait child: exit status 2\n",
		},
		{
			"/usr/bin/false child status 1 is forwarded",
			[]string{"false"},
			1,
			"timeit: wait child: exit status 1\n",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var gotOut bytes.Buffer
			gotCode := run(PROGNAME, tc.args, &gotOut)
			if gotCode != tc.wantCode {
				t.Errorf("\ncode: got: %d; want: %d", gotCode, tc.wantCode)
			}
			if gotOut.String() != tc.wantOut {
				t.Errorf("\noutput: got: %q; want: %q", gotOut.String(), tc.wantOut)
			}
		})
	}

}
