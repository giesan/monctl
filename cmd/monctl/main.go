package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/giesan/monitorcontrol/internal/ddc"
	"github.com/giesan/monitorcontrol/internal/monitor"
)

const usage = "Usage: monctl list | get <bus> | set <bus> <percent>"

func main() {
	os.Exit(run(context.Background(), os.Args[1:], monitor.Service{Backend: ddc.New()}, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, service monitor.Service, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if (args[0] == "list" && len(args) != 1) || (args[0] == "get" && len(args) != 2) || (args[0] == "set" && len(args) != 3) {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
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
		bus, parseErr := strconv.Atoi(args[1])
		if parseErr != nil || bus < 0 {
			fmt.Fprintln(stderr, "bus must be a non-negative number")
			fmt.Fprintln(stderr, usage)
			return 2
		}
		if args[0] == "get" {
			var percent int
			percent, err = service.Get(ctx, bus)
			if err == nil {
				fmt.Fprintf(stdout, "%d%%\n", percent)
			}
		} else {
			var percent int
			percent, parseErr = strconv.Atoi(args[2])
			if parseErr != nil || percent < 0 || percent > 100 {
				fmt.Fprintln(stderr, "brightness must be an integer from 0 to 100")
				fmt.Fprintln(stderr, usage)
				return 2
			}
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
