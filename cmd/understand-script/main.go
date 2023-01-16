package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(Main())
}

// Main does what unfortunately many libraries do:
//   - call os.Exit() directly
//   - read os.Args directly
//   - write to os.Stdout/os.Stderr
func Main() int {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "return-0":
			fmt.Println(os.Args[1])
			return 0
		case "return-2":
			fmt.Println(os.Args[1])
			return 2
		case "exit-11":
			fmt.Println(os.Args[1])
			os.Exit(11)
		}
	}
	fmt.Fprint(os.Stderr, "usage: understand-script return-0 | return-2 | exit-11\n")
	os.Exit(42)

	return -1 // impossible, make linter happy :-/
}
