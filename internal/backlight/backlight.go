// Package backlight controls a laptop panel through Linux backlight sysfs and logind.
package backlight

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const logind = "org.freedesktop.login1"
const managerPath = "/org/freedesktop/login1"

type runner interface {
	Run(context.Context, ...string) (string, error)
}

type executable struct{}

func (executable) Run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "busctl", args...).CombinedOutput()
	return string(out), err
}

type Controller struct {
	root      string
	sessionID func() string
	pid       func() int
	run       runner
}

func New() *Controller {
	return &Controller{
		root:      "/sys/class/backlight",
		sessionID: func() string { return os.Getenv("XDG_SESSION_ID") },
		pid:       os.Getpid,
		run:       executable{},
	}
}

func (c *Controller) device() (string, error) {
	entries, err := os.ReadDir(c.root)
	if err != nil {
		return "", fmt.Errorf("list laptop backlight devices: %w", err)
	}
	var devices []string
	for _, entry := range entries {
		info, err := os.Stat(filepath.Join(c.root, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("inspect backlight %s: %w", entry.Name(), err)
		}
		if info.IsDir() {
			devices = append(devices, entry.Name())
		}
	}
	switch len(devices) {
	case 0:
		return "", fmt.Errorf("no laptop backlight found under %s", c.root)
	case 1:
		return devices[0], nil
	default:
		return "", fmt.Errorf("multiple laptop backlights found (%s); automatic selection is ambiguous", strings.Join(devices, ", "))
	}
}

func readNumber(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 32)
}

func (c *Controller) level() (string, uint64, uint64, error) {
	name, err := c.device()
	if err != nil {
		return "", 0, 0, err
	}
	base := filepath.Join(c.root, name)
	max, err := readNumber(filepath.Join(base, "max_brightness"))
	if err != nil {
		return "", 0, 0, fmt.Errorf("read maximum brightness of %s: %w", name, err)
	}
	current, err := readNumber(filepath.Join(base, "brightness"))
	if err != nil {
		return "", 0, 0, fmt.Errorf("read brightness of %s: %w", name, err)
	}
	if max == 0 || current > max {
		return "", 0, 0, fmt.Errorf("invalid laptop brightness %d/%d on %s", current, max, name)
	}
	return name, current, max, nil
}

func (c *Controller) Get(context.Context) (int, error) {
	_, current, max, err := c.level()
	if err != nil {
		return 0, err
	}
	return int((current*100 + max/2) / max), nil
}

var sessionPath = regexp.MustCompile(`^/org/freedesktop/login1/session/[A-Za-z0-9_]+$`)

func (c *Controller) session(ctx context.Context) (string, error) {
	args := []string{"--system", "call", logind, managerPath, logind + ".Manager"}
	if id := c.sessionID(); id != "" {
		args = append(args, "GetSession", "s", id)
	} else {
		args = append(args, "GetSessionByPID", "u", strconv.Itoa(c.pid()))
	}
	out, err := c.run.Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("find graphical login session via logind: %s: %w", strings.TrimSpace(out), err)
	}
	fields := strings.Fields(out)
	if len(fields) != 2 || fields[0] != "o" {
		return "", fmt.Errorf("unexpected logind session response: %q", strings.TrimSpace(out))
	}
	path := strings.Trim(fields[1], `"`)
	if !sessionPath.MatchString(path) {
		return "", fmt.Errorf("invalid logind session path: %q", path)
	}
	return path, nil
}

func (c *Controller) Set(ctx context.Context, percent int) error {
	if percent < 0 || percent > 100 {
		return fmt.Errorf("laptop brightness must be between 0 and 100 percent")
	}
	name, _, max, err := c.level()
	if err != nil {
		return err
	}
	value := (uint64(percent)*max + 50) / 100
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	path, err := c.session(ctx)
	if err != nil {
		return err
	}
	out, err := c.run.Run(ctx, "--system", "call", logind, path, logind+".Session", "SetBrightness", "ssu", "backlight", name, strconv.FormatUint(value, 10))
	if err != nil {
		return fmt.Errorf("set laptop brightness via logind: %s: %w", strings.TrimSpace(out), err)
	}
	return nil
}
