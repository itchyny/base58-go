package cli

import "os"

// Run base58
func Run() int {
	return (&cli{}).run(os.Args[1:])
}
