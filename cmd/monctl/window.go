//go:build !nucular_shiny

package main

import (
	"fmt"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

func runWindow(monitor.Service, laptopControl) error {
	return fmt.Errorf("window unavailable; build with -tags=nucular_shiny (CGO_ENABLED=0)")
}
