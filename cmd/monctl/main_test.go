package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type backend struct {
	write int
}

func (b *backend) Detect(context.Context) ([]monitor.Display, error) {
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
		{[]string{"set", "5", "40"}, 0, "Requested 40%"},
		{[]string{"set", "5", "101"}, 2, ""},
		{[]string{"get", "x"}, 2, ""},
		{[]string{"list", "extra"}, 2, ""},
	} {
		var out, errs bytes.Buffer
		code := run(context.Background(), tc.args, monitor.Service{Backend: b}, &out, &errs)
		if code != tc.code || !strings.Contains(out.String(), tc.output) {
			t.Errorf("args %v: exit=%d stdout=%q stderr=%q", tc.args, code, out.String(), errs.String())
		}
	}
	if b.write != 20 {
		t.Fatalf("write = %d, want 20", b.write)
	}
}
