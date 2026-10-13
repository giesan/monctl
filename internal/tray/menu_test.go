package tray

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/godbus/dbus/v5"
)

type fakeMonitors struct {
	mu       sync.Mutex
	displays []monitor.Display
	levels   map[int]int
}

func (f *fakeMonitors) List(context.Context) ([]monitor.Display, error) { return f.displays, nil }
func (f *fakeMonitors) Get(_ context.Context, bus int) (int, error) {
	return f.level(bus), nil
}
func (f *fakeMonitors) level(bus int) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.levels[bus]
}
func (f *fakeMonitors) Set(_ context.Context, bus, value int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.levels[bus] = value
	return nil
}

type fakeLaptop struct{ value int }

func (f *fakeLaptop) Get(context.Context) (int, error) { return f.value, nil }
func (f *fakeLaptop) Set(_ context.Context, value int) error {
	f.value = value
	return nil
}

func TestMenuActionsUseCorrectDisplayAndClamp(t *testing.T) {
	monitors := &fakeMonitors{displays: []monitor.Display{{Bus: 7, Model: "PHL"}, {Bus: 9, Model: "DELL"}}, levels: map[int]int{7: 9, 9: 99}}
	laptop := &fakeLaptop{value: 2}
	m := model{monitors: monitors, laptop: laptop}
	m.refresh(context.Background())
	if len(m.entries) != 11 {
		t.Fatalf("menu entries = %d, want 11", len(m.entries))
	}
	if !strings.Contains(m.entries[0].label, "PHL (bus 7): 9%") || !strings.Contains(m.entries[6].label, "Laptop: 2%") {
		t.Fatalf("unexpected menu: %+v", m.entries)
	}
	for _, id := range []int32{2, 6, 8} {
		if err := m.apply(context.Background(), m.actions[id]); err != nil {
			t.Fatal(err)
		}
	}
	if monitors.levels[7] != 4 || monitors.levels[9] != 100 || laptop.value != 0 {
		t.Fatalf("updated brightness: displays=%v laptop=%d", monitors.levels, laptop.value)
	}
}

func TestDBusMenuSignatures(t *testing.T) {
	if got := dbus.SignatureOf(menuLayout{}).String(); got != "(ia{sv}av)" {
		t.Fatalf("GetLayout signature = %s", got)
	}
	if got := dbus.SignatureOf([]menuProperty{}).String(); got != "a(ia{sv})" {
		t.Fatalf("GetGroupProperties signature = %s", got)
	}
	if got := dbus.SignatureOf([]menuEvent{}).String(); got != "a(isvu)" {
		t.Fatalf("EventGroup signature = %s", got)
	}
	if got := dbus.SignatureOf([]pixmap{}).String(); got != "a(iiay)" {
		t.Fatalf("icon signature = %s", got)
	}
	if got := dbus.SignatureOf(tooltip{}).String(); got != "(sa(iiay)ss)" {
		t.Fatalf("tooltip signature = %s", got)
	}
}
