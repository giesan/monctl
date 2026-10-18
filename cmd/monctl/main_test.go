package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/giesan/monitorcontrol/internal/version"
)

type backend struct {
	write    int
	displays []monitor.Display
	feature  monitor.Feature
}

type fakeSetup struct{ applied bool }

type fakeLaptop struct{ set int }

func (*fakeLaptop) Get(context.Context) (int, error) { return 75, nil }
func (f *fakeLaptop) Set(_ context.Context, percent int) error {
	f.set = percent
	return nil
}

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
func (b *backend) Read(context.Context, int, monitor.Feature) (monitor.Level, error) {
	return monitor.Level{Current: 25, Maximum: 50}, nil
}
func (b *backend) Write(_ context.Context, _ int, feature monitor.Feature, value int) error {
	b.write = value
	b.feature = feature
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
		{[]string{"contrast", "get"}, 0, "50%"},
		{[]string{"contrast", "set", "40"}, 0, "Requested 40% contrast"},
		{[]string{"volume", "get", "5"}, 0, "50%"},
		{[]string{"volume", "set", "5", "40"}, 0, "Requested 40% monitor volume"},
		{[]string{"set", "5", "101"}, 2, ""},
		{[]string{"volume", "set", "101"}, 2, ""},
		{[]string{"volume", "oops"}, 2, ""},
		{[]string{"set", "101"}, 2, ""},
		{[]string{"get", "x"}, 2, ""},
		{[]string{"list", "extra"}, 2, ""},
	} {
		var out, errs bytes.Buffer
		code := run(context.Background(), tc.args, monitor.Service{Backend: b}, nil, nil, &out, &errs)
		if code != tc.code || !strings.Contains(out.String(), tc.output) {
			t.Errorf("args %v: exit=%d stdout=%q stderr=%q", tc.args, code, out.String(), errs.String())
		}
	}
	if b.write != 20 || b.feature != monitor.Volume {
		t.Fatalf("write = %d, feature = %s; want volume 20", b.write, b.feature)
	}
}

func TestWindowOnlyOpensWithoutArguments(t *testing.T) {
	b := &backend{}
	opened := 0
	open := func() error {
		opened++
		return nil
	}
	var out, errs bytes.Buffer
	if code := runApp(context.Background(), nil, monitor.Service{Backend: b}, nil, nil, open, &out, &errs); code != 0 || opened != 1 || out.Len() != 0 || errs.Len() != 0 {
		t.Fatalf("window: exit=%d opened=%d stdout=%q stderr=%q", code, opened, out.String(), errs.String())
	}
	if code := runApp(context.Background(), []string{"list"}, monitor.Service{Backend: b}, nil, nil, open, &out, &errs); code != 0 || opened != 1 || !strings.Contains(out.String(), "Example") {
		t.Fatalf("list: exit=%d opened=%d stdout=%q stderr=%q", code, opened, out.String(), errs.String())
	}
	if code := runApp(context.Background(), []string{"unknown"}, monitor.Service{Backend: b}, nil, nil, open, &out, &errs); code != 2 || opened != 1 || !strings.Contains(errs.String(), usage) {
		t.Fatalf("unknown: exit=%d opened=%d stderr=%q", code, opened, errs.String())
	}
}

func TestVersionDoesNotOpenWindowOrDetectMonitors(t *testing.T) {
	opened := false
	var out, errs bytes.Buffer
	code := runApp(context.Background(), []string{"--version"}, monitor.Service{}, nil, nil, func() error {
		opened = true
		return nil
	}, &out, &errs)
	if code != 0 || opened || errs.Len() != 0 || out.String() != "monctl "+version.Value+"\n" {
		t.Fatalf("version: exit=%d opened=%t stdout=%q stderr=%q", code, opened, out.String(), errs.String())
	}
	if code := runApp(context.Background(), []string{"--version", "extra"}, monitor.Service{}, nil, nil, nil, &out, &errs); code != 2 {
		t.Fatalf("extra version argument: exit=%d", code)
	}
}

func TestWindowError(t *testing.T) {
	var out, errs bytes.Buffer
	code := runApp(context.Background(), nil, monitor.Service{}, nil, nil, func() error {
		return fmt.Errorf("no X11 display available")
	}, &out, &errs)
	if code != 1 || out.Len() != 0 || !strings.Contains(errs.String(), "monctl: no X11 display available") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out.String(), errs.String())
	}
}

func TestImplicitMonitorFailsForMultipleDisplays(t *testing.T) {
	b := &backend{displays: []monitor.Display{{Bus: 5, Model: "A"}, {Bus: 6, Model: "B"}}}
	var out, errs bytes.Buffer
	code := run(context.Background(), []string{"set", "50"}, monitor.Service{Backend: b}, nil, nil, &out, &errs)
	if code != 1 || b.write != 0 || !strings.Contains(errs.String(), "buses 5, 6") {
		t.Fatalf("exit=%d write=%d stdout=%q stderr=%q", code, b.write, out.String(), errs.String())
	}
	code = run(context.Background(), []string{"get"}, monitor.Service{Backend: b}, nil, nil, &out, &errs)
	if code != 1 {
		t.Fatalf("ambiguous get: exit=%d", code)
	}
}

func TestSetupAndPrivilegedHelper(t *testing.T) {
	b := &backend{}
	s := &fakeSetup{}
	var out, errs bytes.Buffer
	if code := run(context.Background(), []string{"setup"}, monitor.Service{Backend: b}, s, nil, &out, &errs); code != 0 || !strings.Contains(out.String(), "1 brightness-capable monitor") {
		t.Fatalf("setup: exit=%d output=%q error=%q", code, out.String(), errs.String())
	}
	if code := run(context.Background(), []string{"setup", "--apply"}, monitor.Service{Backend: b}, s, nil, &out, &errs); code != 0 || !s.applied {
		t.Fatalf("apply: exit=%d applied=%t error=%q", code, s.applied, errs.String())
	}
	if code := run(context.Background(), []string{"setup", "bad"}, monitor.Service{Backend: b}, s, nil, &out, &errs); code != 2 {
		t.Fatalf("bad args: exit=%d", code)
	}
}

func TestLaptopCommands(t *testing.T) {
	laptop := &fakeLaptop{}
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"laptop", "get"}, 0, "75%"},
		{[]string{"laptop", "set", "40"}, 0, "Requested 40%"},
		{[]string{"laptop", "set", "101"}, 2, ""},
		{[]string{"laptop", "set"}, 2, ""},
	} {
		var out, errs bytes.Buffer
		code := run(context.Background(), tc.args, monitor.Service{}, nil, laptop, &out, &errs)
		if code != tc.code || !strings.Contains(out.String(), tc.want) {
			t.Errorf("args=%v code=%d out=%q err=%q", tc.args, code, out.String(), errs.String())
		}
	}
	if laptop.set != 40 {
		t.Fatalf("set value = %d", laptop.set)
	}
}
