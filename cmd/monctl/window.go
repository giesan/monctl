//go:build !gtk

package main

import (
	"fmt"

	"github.com/giesan/monctl/internal/monitor"
)

func runWindow(monitor.Service, laptopControl) error {
	return fmt.Errorf("window unavailable; build with -tags=gtk and CGO_ENABLED=1 (GTK 3 development files required)")
}
