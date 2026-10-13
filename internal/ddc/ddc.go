// Package ddc talks to external monitors through the ddcutil executable.
package ddc

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

const commandTimeout = 20 * time.Second

type runner interface {
	Run(context.Context, ...string) (string, error)
}

type executable struct{}

func (executable) Run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "ddcutil", args...).CombinedOutput()
	return string(out), err
}

type Backend struct {
	run runner
	mu  sync.Mutex
}

func New() *Backend { return &Backend{run: executable{}} }

// NewWithRunner is intentionally not exposed: production commands always use ddcutil.
func (b *Backend) command(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	// Serialize transactions within this process; ddcutil also coordinates across processes.
	b.mu.Lock()
	defer b.mu.Unlock()
	out, err := b.run.Run(ctx, args...)
	if err == nil {
		return out, nil
	}
	message := strings.TrimSpace(out)
	low := strings.ToLower(message)
	switch {
	case errors.Is(err, exec.ErrNotFound):
		return "", fmt.Errorf("ddcutil not found; install it with your distribution's package manager: %w", err)
	case ctx.Err() != nil:
		return "", fmt.Errorf("ddcutil timed out or was cancelled: %w", ctx.Err())
	case strings.Contains(low, "no /dev/i2c devices"):
		return "", fmt.Errorf("no /dev/i2c-* devices found; load the i2c-dev kernel module (sudo modprobe i2c-dev) and check your graphics driver: %w", err)
	case strings.Contains(low, "permission denied"), strings.Contains(low, "eacces"):
		return "", fmt.Errorf("cannot access /dev/i2c-*; check I2C device permissions: %s: %w", message, err)
	case strings.Contains(low, "unsupported"), strings.Contains(low, "not supported"), strings.Contains(low, "vcp 10 err"):
		return "", fmt.Errorf("%w: %s", monitor.ErrUnsupported, message)
	case strings.Contains(low, "no displays found"), strings.Contains(low, "no monitor"), strings.Contains(low, "not found"):
		return "", fmt.Errorf("monitor not found; rerun monctl list and check that DDC/CI is enabled: %s: %w", message, err)
	default:
		return "", fmt.Errorf("ddcutil %s failed (check DDC/CI in the monitor menu): %s: %w", args[0], message, err)
	}
}

var busLine = regexp.MustCompile(`(?m)^\s*I2C bus:\s*/dev/i2c-([0-9]+)\s*$`)
var monitorLine = regexp.MustCompile(`(?m)^\s*Monitor:\s*[^:\r\n]*:([^:\r\n]+):[^\r\n]*$`)
var displayHeader = regexp.MustCompile(`(?m)^Display [0-9]+\s*$`)

func (b *Backend) Detect(ctx context.Context) ([]monitor.Display, error) {
	out, err := b.command(ctx, "detect", "--brief")
	if err != nil {
		return nil, err
	}
	indices := displayHeader.FindAllStringIndex(out, -1)
	if len(indices) == 0 && strings.TrimSpace(out) != "" {
		return nil, fmt.Errorf("unexpected ddcutil detect output: %q", strings.TrimSpace(out))
	}
	displays := make([]monitor.Display, 0, len(indices))
	for i, index := range indices {
		end := len(out)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		block := out[index[1]:end]
		bus := busLine.FindStringSubmatch(block)
		model := monitorLine.FindStringSubmatch(block)
		if bus == nil || model == nil {
			return nil, fmt.Errorf("unrecognized DDC monitor entry: %q", strings.TrimSpace(block))
		}
		n, err := strconv.Atoi(bus[1])
		if err != nil {
			return nil, fmt.Errorf("invalid I2C bus %q: %w", bus[1], err)
		}
		displays = append(displays, monitor.Display{Bus: n, Model: strings.TrimSpace(model[1])})
	}
	return displays, nil
}

var vcpLine = regexp.MustCompile(`(?m)^VCP 10 C ([0-9]+) ([0-9]+)\s*$`)

func (b *Backend) Read(ctx context.Context, bus int) (monitor.Brightness, error) {
	out, err := b.command(ctx, "getvcp", "10", "--bus", strconv.Itoa(bus), "--terse")
	if err != nil {
		return monitor.Brightness{}, err
	}
	match := vcpLine.FindStringSubmatch(out)
	if match == nil {
		if strings.Contains(out, "VCP 10 ERR") {
			return monitor.Brightness{}, monitor.ErrUnsupported
		}
		return monitor.Brightness{}, fmt.Errorf("unexpected ddcutil brightness response: %q", strings.TrimSpace(out))
	}
	current, err := strconv.Atoi(match[1])
	if err != nil {
		return monitor.Brightness{}, fmt.Errorf("invalid current brightness: %w", err)
	}
	maximum, err := strconv.Atoi(match[2])
	if err != nil {
		return monitor.Brightness{}, fmt.Errorf("invalid maximum brightness: %w", err)
	}
	if maximum <= 0 || current > maximum {
		return monitor.Brightness{}, fmt.Errorf("monitor reported invalid brightness %d/%d", current, maximum)
	}
	return monitor.Brightness{Current: current, Maximum: maximum}, nil
}

func (b *Backend) Write(ctx context.Context, bus, value int) error {
	_, err := b.command(ctx, "setvcp", "10", strconv.Itoa(value), "--bus", strconv.Itoa(bus))
	return err
}
