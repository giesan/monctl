package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/giesan/monitorcontrol/internal/backlight"
	"github.com/giesan/monitorcontrol/internal/ddc"
	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/tray"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server, err := tray.Start(ctx, monitor.Service{Backend: ddc.New()}, backlight.New())
	if err == nil {
		err = server.Wait(ctx)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "monctl-tray:", err)
		os.Exit(1)
	}
}
