package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/giesan/monitorcontrol/internal/ddc"
	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/setup"
)

const usage = "Usage: monctl setup | list | get [bus] | set [bus] <percent>"

func main() {
	configuration, err := setup.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "monctl:", err)
		os.Exit(1)
	}
	os.Exit(run(context.Background(), os.Args[1:], monitor.Service{Backend: ddc.New()}, configuration, os.Stdout, os.Stderr))
}

type setupCommand interface {
	Run(context.Context, io.Writer, io.Writer) error
	Apply(context.Context) error
}

func run(ctx context.Context, args []string, service monitor.Service, configuration setupCommand, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if (args[0] == "setup" && len(args) != 1 && !(len(args) == 2 && args[1] == "--apply")) ||
		(args[0] == "list" && len(args) != 1) ||
		(args[0] == "get" && len(args) > 2) ||
		(args[0] == "set" && (len(args) < 2 || len(args) > 3)) {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "setup":
		if len(args) == 2 {
			err = configuration.Apply(ctx)
			break
		}
		if err = configuration.Run(ctx, stdout, stderr); err != nil {
			break
		}
		var displays []monitor.Display
		displays, err = service.List(ctx)
		if err == nil {
			fmt.Fprintf(stdout, "DDC/CI check: %d brightness-capable monitor(s) detected.\n", len(displays))
		}
	case "list":
		var displays []monitor.Display
		displays, err = service.List(ctx)
		if err == nil {
			if len(displays) == 0 {
				fmt.Fprintln(stdout, "No DDC/CI brightness-capable monitors found.")
			}
			for _, display := range displays {
				fmt.Fprintf(stdout, "%d\t%s\n", display.Bus, display.Model)
			}
		}
	case "get", "set":
		var bus int
		explicitBus := (args[0] == "get" && len(args) == 2) || (args[0] == "set" && len(args) == 3)
		if explicitBus {
			var parseErr error
			bus, parseErr = strconv.Atoi(args[1])
			if parseErr != nil || bus < 0 {
				fmt.Fprintln(stderr, "bus must be a non-negative number")
				fmt.Fprintln(stderr, usage)
				return 2
			}
		}
		var percent int
		if args[0] == "set" {
			value := args[len(args)-1]
			var parseErr error
			percent, parseErr = strconv.Atoi(value)
			if parseErr != nil || percent < 0 || percent > 100 {
				fmt.Fprintln(stderr, "brightness must be an integer from 0 to 100")
				fmt.Fprintln(stderr, usage)
				return 2
			}
		}
		if !explicitBus {
			var display monitor.Display
			display, err = service.Single(ctx)
			if err != nil {
				break
			}
			bus = display.Bus
		}
		if args[0] == "get" {
			percent, err = service.Get(ctx, bus)
			if err == nil {
				fmt.Fprintf(stdout, "%d%%\n", percent)
			}
		} else {
			err = service.Set(ctx, bus, percent)
			if err == nil {
				fmt.Fprintf(stdout, "Requested %d%% brightness on bus %d.\n", percent, bus)
			}
		}
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, "monctl:", err)
		return 1
	}
	return 0
}
