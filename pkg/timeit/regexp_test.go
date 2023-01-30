package timeit

import (
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestUnderstandRegexp(t *testing.T) {
	type match struct {
		input string
		want  []string
	}

	type testCase struct {
		name    string
		re      string
		matches []match
	}

	run := func(t *testing.T, tc testCase) {
		re := regexp.MustCompile(tc.re)
		for _, m := range tc.matches {
			have := re.FindStringSubmatch(m.input)
			assert.Check(t, cmp.DeepEqual(have, m.want), "input: %q", m.input)
		}
	}

	testCases := []testCase{
		{
			name: "want only numbers",
			//    1        2        3
			re: `^([0-9]+) ([0-9]+) ([0-9]+)$`,
			matches: []match{
				{
					input: "1 12 123",
					//             0           1    2     3
					want: []string{"1 12 123", "1", "12", "123"},
				},
				{input: "a ab abc", want: nil},
				{input: "1 12 abc", want: nil},
			},
		},
		{
			name: "want only words",
			//    1        2        3
			re: `^([a-z]+) ([a-z]+) ([a-z]+)$`,
			matches: []match{
				{input: "1 12 123", want: nil},
				{
					input: "a ab abc",
					//             0           1    2     3
					want: []string{"a ab abc", "a", "ab", "abc"},
				},
				{input: "1 12 abc", want: nil},
			},
		},
		{
			name: "want both",
			//    1--------------------------- 5---------------------------
			//     2        3        4          6        7        8
			re: `^(([0-9]+) ([0-9]+) ([0-9]+))|(([a-z]+) ([a-z]+) ([a-z]+))$`,
			matches: []match{
				{
					input: "1 12 123",
					//             0           1           2    3     4
					want: []string{"1 12 123", "1 12 123", "1", "12", "123", "", "", "", ""},
				},
				{
					input: "a ab abc",
					//             0           1   2   3   4   5           6    7     8
					want: []string{"a ab abc", "", "", "", "", "a ab abc", "a", "ab", "abc"},
				},
				{input: "1 12 abc", want: nil},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestUnderstandRegexpNamedGroups(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  map[string]string
	}

	// (?:re)        non-capturing group
	// (?P<name>re)  named and numbered capturing group

	pat := `(?:(?P<A>A:\d+) (?P<B>B:[a-z]+) (?P<C>C:\d+) )?(?P<D>D:[A-Z]+)`
	re := regexp.MustCompile(pat)
	groupNames := re.SubexpNames()

	run := func(t *testing.T, tc testCase) {
		matches := re.FindStringSubmatch(tc.input)
		var groups map[string]string
		if matches != nil {
			groups = make(map[string]string, len(groupNames))
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					groups[groupNames[i]] = matches[i]
				}
			}
		}
		assert.Check(t, cmp.DeepEqual(groups, tc.want), "input: %q", tc.input)
	}

	testCases := []testCase{
		{
			name:  "match case 1 start",
			input: "D:X",
			want: map[string]string{
				"D": "D:X",
			},
		},
		{
			name:  "match case 2 finish",
			input: "A:1 B:a C:2 D:X",
			want: map[string]string{
				"A": "A:1",
				"B": "B:a",
				"C": "C:2",
				"D": "D:X",
			},
		},
		{
			name:  "no match",
			input: "A:1 B:a C:2 D:3",
			want:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { run(t, tc) })
	}
}

func TestPytestRegexp(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  map[string]string
	}

	groupNames := pytestRe.SubexpNames()

	run := func(t *testing.T, tc testCase) {
		matches := pytestRe.FindStringSubmatch(tc.input)
		var groups map[string]string
		if matches != nil {
			groups = make(map[string]string, len(groupNames))
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					groups[groupNames[i]] = matches[i]
				}
			}
		}

		assert.Check(t, cmp.DeepEqual(groups, tc.want), "input: %q", tc.input)
	}

	testCases := []testCase{
		{name: "no match 1", input: "some output that is not a test name...", want: nil},
		{name: "no match 2", input: "this should::not be matched", want: nil},
		{name: "no match 3", input: "thisshouldnot::bematched", want: nil},
		{
			name:  "start of test",
			input: "test_fruits.py::test_bananas",
			want: map[string]string{
				"name": "test_fruits.py::test_bananas",
			},
		},
		{
			name:  "end of test, FAILED, less than 100% has space",
			input: "[gw3] [ 80%] FAILED test_fruits.py::test_apples",
			want: map[string]string{
				"gw":     "[gw3]",
				"name":   "test_fruits.py::test_apples",
				"pct":    "[ 80%]",
				"status": "FAILED",
			},
		},
		{
			name:  "end of test, PASSED, less than 100% has space",
			input: "[gw3] [ 10%] PASSED test_fruits.py::test_apples",
			want: map[string]string{
				"gw":     "[gw3]",
				"name":   "test_fruits.py::test_apples",
				"pct":    "[ 10%]",
				"status": "PASSED",
			},
		},
		{
			name:  "end of test, 100% has no space",
			input: "[gw4] [100%] PASSED test_fruits.py::test_apples",
			want: map[string]string{
				"gw":     "[gw4]",
				"name":   "test_fruits.py::test_apples",
				"pct":    "[100%]",
				"status": "PASSED",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
