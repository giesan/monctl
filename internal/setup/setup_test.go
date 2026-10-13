package setup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	commands [][]string
	err      error
}

func (f *fakeRunner) Run(_ context.Context, _ io.Reader, _, _ io.Writer, name string, args ...string) error {
	f.commands = append(f.commands, append([]string{name}, args...))
	return f.err
}

func fixture(t *testing.T, euid int) (*Setup, *fakeRunner) {
	t.Helper()
	dir := t.TempDir()
	f := &fakeRunner{}
	return &Setup{
		configPath: filepath.Join(dir, "i2c-dev.conf"),
		modulePath: filepath.Join(dir, "i2c_dev"),
		executable: "/test/monctl",
		euid:       func() int { return euid },
		run:        f,
	}, f
}

func TestAlreadyConfiguredNeedsNoSudo(t *testing.T) {
	s, f := fixture(t, 1000)
	if err := os.Mkdir(s.modulePath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.configPath, []byte(configuration), 0644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := s.Run(context.Background(), &out, &out); err != nil || len(f.commands) != 0 {
		t.Fatalf("run: output=%q err=%v commands=%v", out.String(), err, f.commands)
	}
}

func TestRunEscalatesOnlyIfNecessary(t *testing.T) {
	s, f := fixture(t, 1000)
	var out bytes.Buffer
	if err := s.Run(context.Background(), &out, &out); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"sudo", "--", "/test/monctl", "setup", "--apply"}}
	if !reflect.DeepEqual(f.commands, want) || !strings.Contains(out.String(), "Will create") {
		t.Fatalf("commands=%v output=%q", f.commands, out.String())
	}
	f.err = errors.New("sudo failed")
	if err := s.Run(context.Background(), &out, &out); err == nil {
		t.Fatal("expected sudo failure")
	}
}

func TestApplyRejectsUnprivilegedAndConflictingFile(t *testing.T) {
	s, f := fixture(t, 1000)
	if err := s.Apply(context.Background()); err == nil || len(f.commands) != 0 {
		t.Fatalf("unprivileged apply: err=%v commands=%v", err, f.commands)
	}
	s.euid = func() int { return 0 }
	if err := os.WriteFile(s.configPath, []byte("unrelated\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(context.Background()); err == nil || len(f.commands) != 0 {
		t.Fatalf("conflicting file: err=%v commands=%v", err, f.commands)
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	s, f := fixture(t, 0)
	if err := s.Apply(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(s.modulePath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil || string(data) != configuration || len(f.commands) != 1 {
		t.Fatalf("config=%q err=%v commands=%v", data, err, f.commands)
	}
	if !reflect.DeepEqual(f.commands[0], []string{"modprobe", "i2c-dev"}) {
		t.Fatalf("command=%v", f.commands[0])
	}
}

func TestFailedModprobeDoesNotWriteConfig(t *testing.T) {
	s, f := fixture(t, 0)
	f.err = errors.New("module unavailable")
	if err := s.Apply(context.Background()); err == nil {
		t.Fatal("expected modprobe failure")
	}
	if _, err := os.Stat(s.configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config should not exist, stat error=%v", err)
	}
}
