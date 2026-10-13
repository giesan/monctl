package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type backend struct {
	write    int
	displays []monitor.Display
}

type fakeSetup struct{ applied bool }

func (f *fakeSetup) Run(_ context.Context, out, _ io.Writer) error {
	fmt.Fprintln(out, "i2c-dev setup complete.")
	return nil
}
func (f *fakeSetup) Apply(context.Context) error {
	f.applied = true
	return nil
}

func (b *backend) Detect(context.Context) ([]monitor.Display, error) {
	if b.displays != nil {
		return b.displays, nil
	}
	return []monitor.Display{{Bus: 5, Model: "Example"}}, nil
}
func (b *backend) Read(context.Context, int) (monitor.Brightness, error) {
	return monitor.Brightness{Current: 25, Maximum: 50}, nil
}
func (b *backend) Write(_ context.Context, _ int, value int) error {
	b.write = value
	return nil
}

func TestCommandsAndInvalidArguments(t *testing.T) {
	b := &backend{}
	for _, tc := range []struct {
		args   []string
		code   int
		output string
	}{
		{[]string{"list"}, 0, "5\tExample"},
		{[]string{"get", "5"}, 0, "50%"},
		{[]string{"get"}, 0, "50%"},
		{[]string{"set", "5", "40"}, 0, "Requested 40%"},
		{[]string{"set", "40"}, 0, "Requested 40%"},
		{[]string{"set", "5", "101"}, 2, ""},
		{[]string{"set", "101"}, 2, ""},
		{[]string{"get", "x"}, 2, ""},
		{[]string{"list", "extra"}, 2, ""},
	} {
		var out, errs bytes.Buffer
		code := run(context.Background(), tc.args, monitor.Service{Backend: b}, nil, &out, &errs)
		if code != tc.code || !strings.Contains(out.String(), tc.output) {
			t.Errorf("args %v: exit=%d stdout=%q stderr=%q", tc.args, code, out.String(), errs.String())
		}
	}
	if b.write != 20 {
		t.Fatalf("write = %d, want 20", b.write)
	}
}

func TestImplicitMonitorFailsForMultipleDisplays(t *testing.T) {
	b := &backend{displays: []monitor.Display{{Bus: 5, Model: "A"}, {Bus: 6, Model: "B"}}}
	var out, errs bytes.Buffer
	code := run(context.Background(), []string{"set", "50"}, monitor.Service{Backend: b}, nil, &out, &errs)
	if code != 1 || b.write != 0 || !strings.Contains(errs.String(), "buses 5, 6") {
		t.Fatalf("exit=%d write=%d stdout=%q stderr=%q", code, b.write, out.String(), errs.String())
	}
	code = run(context.Background(), []string{"get"}, monitor.Service{Backend: b}, nil, &out, &errs)
	if code != 1 {
		t.Fatalf("ambiguous get: exit=%d", code)
	}
}

func TestSetupAndPrivilegedHelper(t *testing.T) {
	b := &backend{}
	s := &fakeSetup{}
	var out, errs bytes.Buffer
	if code := run(context.Background(), []string{"setup"}, monitor.Service{Backend: b}, s, &out, &errs); code != 0 || !strings.Contains(out.String(), "1 brightness-capable monitor") {
		t.Fatalf("setup: exit=%d output=%q error=%q", code, out.String(), errs.String())
	}
	if code := run(context.Background(), []string{"setup", "--apply"}, monitor.Service{Backend: b}, s, &out, &errs); code != 0 || !s.applied {
		t.Fatalf("apply: exit=%d applied=%t error=%q", code, s.applied, errs.String())
	}
	if code := run(context.Background(), []string{"setup", "bad"}, monitor.Service{Backend: b}, s, &out, &errs); code != 2 {
		t.Fatalf("bad args: exit=%d", code)
	}
}
