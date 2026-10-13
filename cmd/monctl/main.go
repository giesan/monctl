package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/giesan/monitorcontrol/internal/backlight"
	"github.com/giesan/monitorcontrol/internal/ddc"
	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/setup"
)

const usage = "Usage: monctl setup | list | get [bus] | set [bus] <percent> | contrast|volume get [bus] | contrast|volume set [bus] <percent> | laptop get | laptop set <percent>"

func main() {
	configuration, err := setup.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "monctl:", err)
		os.Exit(1)
	}
	os.Exit(run(context.Background(), os.Args[1:], monitor.Service{Backend: ddc.New()}, configuration, backlight.New(), os.Stdout, os.Stderr))
}

type setupCommand interface {
	Run(context.Context, io.Writer, io.Writer) error
	Apply(context.Context) error
}

type laptopControl interface {
	Get(context.Context) (int, error)
	Set(context.Context, int) error
}

func run(ctx context.Context, args []string, service monitor.Service, configuration setupCommand, laptop laptopControl, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if (args[0] == "setup" && len(args) != 1 && !(len(args) == 2 && args[1] == "--apply")) ||
		(args[0] == "list" && len(args) != 1) {
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
		return runControl(ctx, args, monitor.Brightness, service, stdout, stderr)
	case "contrast", "volume":
		feature := monitor.Contrast
		if args[0] == "volume" {
			feature = monitor.Volume
		}
		return runControl(ctx, args[1:], feature, service, stdout, stderr)
	case "laptop":
		return runLaptop(ctx, args[1:], laptop, stdout, stderr)
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

func runLaptop(ctx context.Context, args []string, laptop laptopControl, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "get" {
		value, err := laptop.Get(ctx)
		if err != nil {
			fmt.Fprintln(stderr, "monctl:", err)
			return 1
		}
		fmt.Fprintf(stdout, "%d%%\n", value)
		return 0
	}
	if len(args) == 2 && args[0] == "set" {
		percent, err := strconv.Atoi(args[1])
		if err != nil || percent < 0 || percent > 100 {
			fmt.Fprintln(stderr, "laptop brightness must be an integer from 0 to 100")
			return 2
		}
		if err = laptop.Set(ctx, percent); err != nil {
			fmt.Fprintln(stderr, "monctl:", err)
			return 1
		}
		fmt.Fprintf(stdout, "Requested %d%% laptop brightness.\n", percent)
		return 0
	}
	fmt.Fprintln(stderr, usage)
	return 2
}

func runControl(ctx context.Context, args []string, feature monitor.Feature, service monitor.Service, stdout, stderr io.Writer) int {
	if len(args) == 0 ||
		(args[0] == "get" && len(args) > 2) ||
		(args[0] == "set" && (len(args) < 2 || len(args) > 3)) ||
		(args[0] != "get" && args[0] != "set") {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	var bus int
	explicitBus := (args[0] == "get" && len(args) == 2) || (args[0] == "set" && len(args) == 3)
	if explicitBus {
		var err error
		bus, err = strconv.Atoi(args[1])
		if err != nil || bus < 0 {
			fmt.Fprintln(stderr, "bus must be a non-negative number")
			fmt.Fprintln(stderr, usage)
			return 2
		}
	}
	var percent int
	if args[0] == "set" {
		var err error
		percent, err = strconv.Atoi(args[len(args)-1])
		if err != nil || percent < 0 || percent > 100 {
			fmt.Fprintln(stderr, feature, "must be an integer from 0 to 100")
			fmt.Fprintln(stderr, usage)
			return 2
		}
	}
	if !explicitBus {
		display, err := service.SingleFeature(ctx, feature)
		if err != nil {
			fmt.Fprintln(stderr, "monctl:", err)
			return 1
		}
		bus = display.Bus
	}
	if args[0] == "get" {
		value, err := service.GetFeature(ctx, bus, feature)
		if err != nil {
			fmt.Fprintln(stderr, "monctl:", err)
			return 1
		}
		fmt.Fprintf(stdout, "%d%%\n", value)
		return 0
	}
	if err := service.SetFeature(ctx, bus, feature, percent); err != nil {
		fmt.Fprintln(stderr, "monctl:", err)
		return 1
	}
	fmt.Fprintf(stdout, "Requested %d%% %s on bus %d.\n", percent, feature, bus)
	return 0
}
