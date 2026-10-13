package window

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type fakeMonitors struct {
	mu     sync.Mutex
	writes []int
}

func (*fakeMonitors) List(context.Context) ([]monitor.Display, error) {
	return []monitor.Display{{Bus: 7, Model: "PHL"}}, nil
}
func (*fakeMonitors) Get(context.Context, int) (int, error) { return 30, nil }
func (f *fakeMonitors) Set(_ context.Context, bus, value int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if bus != 7 {
		return nil
	}
	f.writes = append(f.writes, value)
	return nil
}

type fakeLaptop struct{ value int }

func (f *fakeLaptop) Get(context.Context) (int, error) { return f.value, nil }
func (f *fakeLaptop) Set(_ context.Context, value int) error {
	f.value = value
	return nil
}

func TestSliderChangesAreDebounced(t *testing.T) {
	monitors := &fakeMonitors{}
	m := New(context.Background(), monitors, &fakeLaptop{value: 70}, nil)
	if len(m.sliders) != 2 || m.sliders[0].label != "PHL (Bus 7)" || m.sliders[1].label != "Laptop" {
		t.Fatalf("sliders: %+v", m.sliders)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(0, 40)
	m.Request(0, 45)
	m.Request(0, 75)
	deadline := time.After(2 * time.Second)
	for {
		monitors.mu.Lock()
		written := append([]int(nil), monitors.writes...)
		monitors.mu.Unlock()
		if len(written) > 0 {
			if len(written) != 1 || written[0] != 75 {
				t.Fatalf("writes = %v, want [75]", written)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("slider did not write")
		case <-time.After(10 * time.Millisecond):
		}
	}
}
