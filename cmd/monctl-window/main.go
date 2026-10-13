//go:build !nucular_shiny

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "monctl-window: build with -tags=nucular_shiny (CGO_ENABLED=0)")
	os.Exit(1)
}
