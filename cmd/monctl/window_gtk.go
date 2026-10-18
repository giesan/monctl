//go:build gtk

package main

import (
	"github.com/giesan/monctl/internal/monitor"
	"github.com/giesan/monctl/internal/window"
)

func runWindow(service monitor.Service, laptop laptopControl) error {
	return window.Run(service, laptop)
}
