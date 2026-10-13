package backlight

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type reply struct {
	out string
	err error
}

type fakeRunner struct {
	replies  []reply
	commands [][]string
}

func (f *fakeRunner) Run(_ context.Context, args ...string) (string, error) {
	f.commands = append(f.commands, args)
	result := f.replies[0]
	f.replies = f.replies[1:]
	return result.out, result.err
}

func fixture(t *testing.T) *Controller {
	t.Helper()
	root := t.TempDir()
	device := filepath.Join(root, "amdgpu_bl1")
	if err := os.Mkdir(device, 0755); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"brightness": "32000\n", "max_brightness": "64000\n"} {
		if err := os.WriteFile(filepath.Join(device, name), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return &Controller{root: root, sessionID: func() string { return "2" }, pid: func() int { return 1234 }}
}

func TestGetAndSetViaLogind(t *testing.T) {
	c := fixture(t)
	f := &fakeRunner{replies: []reply{{out: "o \"/org/freedesktop/login1/session/_32\"\n"}, {}}}
	c.run = f
	value, err := c.Get(context.Background())
	if err != nil || value != 50 {
		t.Fatalf("get=%d err=%v", value, err)
	}
	if err := c.Set(context.Background(), 75); err != nil {
		t.Fatal(err)
	}
	want := []string{"--system", "call", logind, "/org/freedesktop/login1/session/_32", logind + ".Session", "SetBrightness", "ssu", "backlight", "amdgpu_bl1", "48000"}
	if len(f.commands) != 2 || !reflect.DeepEqual(f.commands[1], want) {
		t.Fatalf("calls = %v", f.commands)
	}
	data, err := os.ReadFile(filepath.Join(c.root, "amdgpu_bl1", "brightness"))
	if err != nil || string(data) != "32000\n" {
		t.Fatalf("sysfs file modified: %q err=%v", data, err)
	}
}

func TestSetRejectsAmbiguousAndInvalidDevice(t *testing.T) {
	c := fixture(t)
	if err := os.Mkdir(filepath.Join(c.root, "intel_backlight"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(context.Background(), 50); err == nil || !strings.Contains(err.Error(), "multiple laptop backlights") {
		t.Fatalf("ambiguous device error = %v", err)
	}
	if err := c.Set(context.Background(), 101); err == nil || !strings.Contains(err.Error(), "between 0 and 100") {
		t.Fatalf("invalid percentage error = %v", err)
	}
}

func TestSetUsesPIDWhenSessionIDMissing(t *testing.T) {
	c := fixture(t)
	c.sessionID = func() string { return "" }
	f := &fakeRunner{replies: []reply{{out: "o \"/org/freedesktop/login1/session/_32\"\n"}, {}}}
	c.run = f
	if err := c.Set(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	want := []string{"--system", "call", logind, managerPath, logind + ".Manager", "GetSessionByPID", "u", "1234"}
	if !reflect.DeepEqual(f.commands[0], want) {
		t.Fatalf("session lookup = %v", f.commands[0])
	}
}

func TestLogindFailureIsReported(t *testing.T) {
	c := fixture(t)
	c.run = &fakeRunner{replies: []reply{{out: "Access denied", err: errors.New("exit status 1")}}}
	if err := c.Set(context.Background(), 50); err == nil || !strings.Contains(err.Error(), "Access denied") {
		t.Fatalf("logind failure = %v", err)
	}
}
