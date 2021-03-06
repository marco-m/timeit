// This code is released under the MIT License
// Copyright (c) 2020 Marco Molteni and the timeit contributors.

package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	SLEEPIT, _ = filepath.Abs("../../bin/sleepit")
)

func TestSmoke(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantOut []string
	}{
		"version": {
			args:    []string{"version"},
			wantOut: []string{"sleepit version "},
		},
		"default": {
			args: []string{"default", "-sleep=1ms"},
			wantOut: []string{
				"sleepit: ready\n",
				"sleepit: work started\n",
				"sleepit: work done\n",
			},
		},
		"handle": {
			args: []string{"handle", "-sleep=1ms"},
			wantOut: []string{
				"sleepit: ready\n",
				"sleepit: work started\n",
				"sleepit: work done\n",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			sut := exec.CommandContext(ctx, SLEEPIT, tc.args...)

			out, err := sut.CombinedOutput()

			if err != nil {
				t.Errorf("exec error: %q; want: no error", err)
			}

			if ctx.Err() != nil {
				t.Errorf("context error: %q", ctx.Err())
			}

			gotLines := strings.SplitAfter(string(out), "\n")
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
