package main

import (
	"fmt"
	"os"

	"github.com/atlasdev/orbitron/internal/license"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "encode" {
		fmt.Fprintln(os.Stderr, "usage: tokenenc encode REAL_TOKEN")
		os.Exit(1)
	}
	fmt.Println(license.EncodeTokenPublic(os.Args[2]))
}
