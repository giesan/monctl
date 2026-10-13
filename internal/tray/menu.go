package tray

import (
	"context"
	"fmt"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type monitors interface {
	List(context.Context) ([]monitor.Display, error)
	Get(context.Context, int) (int, error)
	Set(context.Context, int, int) error
}

type laptop interface {
	Get(context.Context) (int, error)
	Set(context.Context, int) error
}

type action struct {
	bus   int
	delta int
	kind  string
}

type entry struct {
	id      int32
	label   string
	enabled bool
}

type model struct {
	monitors monitors
	laptop   laptop
	entries  []entry
	actions  map[int32]action
}

func (m *model) refresh(ctx context.Context) {
	m.entries = nil
	m.actions = make(map[int32]action)
	add := func(label string, enabled bool, act action) {
		id := int32(len(m.entries) + 1)
		m.entries = append(m.entries, entry{id: id, label: label, enabled: enabled})
		if enabled {
			m.actions[id] = act
		}
	}
	displays, err := m.monitors.List(ctx)
	if err != nil {
		add("External monitors: "+err.Error(), false, action{})
	}
	for _, display := range displays {
		value, err := m.monitors.Get(ctx, display.Bus)
		if err != nil {
			add(display.Model+": "+err.Error(), false, action{})
			continue
		}
		add(fmt.Sprintf("%s (bus %d): %d%%", display.Model, display.Bus, value), false, action{})
		add("  − 5%", true, action{kind: "external", bus: display.Bus, delta: -5})
		add("  + 5%", true, action{kind: "external", bus: display.Bus, delta: 5})
	}
	if value, err := m.laptop.Get(ctx); err == nil {
		add(fmt.Sprintf("Laptop: %d%%", value), false, action{})
		add("  − 5%", true, action{kind: "laptop", delta: -5})
		add("  + 5%", true, action{kind: "laptop", delta: 5})
	}
	if len(m.entries) == 0 {
		add("Keine Helligkeitsregler gefunden", false, action{})
	}
	add("Schieberegler öffnen", true, action{kind: "window"})
	add("Beenden", true, action{kind: "quit"})
}

func (m *model) apply(ctx context.Context, act action) error {
	var current int
	var err error
	switch act.kind {
	case "external":
		current, err = m.monitors.Get(ctx, act.bus)
	case "laptop":
		current, err = m.laptop.Get(ctx)
	default:
		return fmt.Errorf("unknown tray action %q", act.kind)
	}
	if err != nil {
		return err
	}
	target := max(0, min(100, current+act.delta))
	if act.kind == "external" {
		return m.monitors.Set(ctx, act.bus, target)
	}
	return m.laptop.Set(ctx, target)
}
