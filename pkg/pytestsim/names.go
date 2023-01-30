// This code is released under the MIT License
// Copyright (c) 2023 Marco Molteni and the timeit contributors.

package pytestsim

import "fmt"

// https://simple.wikipedia.org/wiki/List_of_vegetables

var (
	items = [][]string{
		// fruits,
		{
			"apple",
			"banana",
			"coconut",
			"grape",
			"mango",
			"papaya",
		},
		// herbs
		{
			"basil",
			"caraway",
			"coriander",
			"dill",
			"lavender",
			"oregano",
			"rosemary",
			"thyme",
		},
		// roots,
		{
			"carrot",
			"chufa",
			"fennel",
			"garlic",
			"ginger",
			"onion",
			"parsley",
			"radish",
		},
	}

	categories = []string{
		"fruits",
		"herbs",
		"roots",
	}
)

// This could be parametrized to emulate different "--dist" behaviors of xdist ...
func names() []string {
	names := make([]string, 0, 50)
	for cat := 0; cat < len(categories); cat++ {
		for it := 0; it < len(items[cat]); it++ {
			names = append(names,
				fmt.Sprintf("test_%s.py::test_%s", categories[cat], items[cat][it]))
		}
	}
	return names
}
