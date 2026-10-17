//go:build nucular_shiny

package main

import (
	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/window"
)

func runWindow(service monitor.Service, laptop laptopControl) error {
	return window.Run(service, laptop)
}
