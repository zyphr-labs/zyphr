// Command zyphr is the Zyphr CLI.
package main

import (
	"fmt"
	"os"

	"github.com/zyphr-labs/zyphr/internal/buildinfo"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("zyphr", buildinfo.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "usage: zyphr version")
	os.Exit(2)
}
