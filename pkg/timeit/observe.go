// This code is released under the MIT License
// Copyright (c) 2023 Marco Molteni and the timeit contributors.

package timeit

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"time"
)

// (?:re)        non-capturing group
// (?P<name>re)  named and numbered capturing group

var pat = `^(?:(?P<gw>\[gw\d+]) +(?P<pct>\[ *\d+%]) +(?P<status>[A-Z]+) +)?(?P<name>.+\.py::.+)$`
var pytestRe = regexp.MustCompile(pat)

func observePytest(records *records, stdout io.Reader, out printFn) {
	scanOut := bufio.NewScanner(stdout)
	groupNames := pytestRe.SubexpNames()
	for scanOut.Scan() {
		line := scanOut.Text()
		now := time.Now()
		fmt.Println(line) // FIXME this can still race...

		matches := pytestRe.FindStringSubmatch(line)

		if matches == nil {
			continue
		}

		// Construct a lookup map for the named groups. In this particular case this is
		// not needed since we look up only one name, but it shows the idiomatic way.
		groups := make(map[string]string, len(groupNames))
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				groups[groupNames[i]] = matches[i]
			}
		}

		// This match is present both for started and landed lines.
		name := groups["name"]

		if groups["status"] == "" {
			// Started.
			records.mu.Lock()
			records.flying[name] = event{name: name, started: now}
			records.mu.Unlock()
		} else {
			// Finished.
			records.mu.Lock()
			evt := records.flying[name]
			evt.finished = now
			delete(records.flying, name)
			records.landed[name] = evt
			records.mu.Unlock()
		}
	}
	if err := scanOut.Err(); err != nil {
		// In this case, we only print the error and keep going.
		out("timeit: reading from stdout: %s\n", err)
	}

	// FIXME WHAT IS THIS FOR???
	// done <- struct{}{}
	// close(done)
}
