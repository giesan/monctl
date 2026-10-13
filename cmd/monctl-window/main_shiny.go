//go:build nucular_shiny

package main

import (
	"fmt"
	"os"

	"github.com/giesan/monitorcontrol/internal/backlight"
	"github.com/giesan/monitorcontrol/internal/ddc"
	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/window"
)

func main() {
	if err := window.Run(monitor.Service{Backend: ddc.New()}, backlight.New()); err != nil {
		fmt.Fprintln(os.Stderr, "monctl-window:", err)
		os.Exit(1)
	}
}
