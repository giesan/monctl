// Package window provides slider state for desktop monitor controls.
package window

import (
	"context"
	"fmt"
	"sync"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type monitors interface {
	Detect(context.Context) ([]monitor.Display, error)
	GetFeature(context.Context, int, monitor.Feature) (int, error)
	SetFeature(context.Context, int, monitor.Feature, int) error
}

type laptop interface {
	Get(context.Context) (int, error)
	Set(context.Context, int) error
}

type slider struct {
	label   string
	value   int // UI goroutine only
	bus     int
	model   string
	feature monitor.Feature
	laptop  bool
	updates chan int
}

type Model struct {
	monitors monitors
	laptop   laptop
	sliders  []*slider
	mu       sync.Mutex
	message  string
	changed  func()
}

func New(ctx context.Context, monitors monitors, laptop laptop, changed func()) *Model {
	m := &Model{monitors: monitors, laptop: laptop, changed: changed}
	displays, err := monitors.Detect(ctx)
	if err != nil {
		m.setMessage(fmt.Sprintf("Externe Monitore: %v", err))
	}
	for _, display := range displays {
		for _, feature := range []monitor.Feature{monitor.Brightness, monitor.Contrast, monitor.Volume} {
			value, err := monitors.GetFeature(ctx, display.Bus, feature)
			if err != nil {
				if !monitor.IsUnsupported(err) {
					m.setMessage(fmt.Sprintf("%s (Bus %d), %s: %v", display.Model, display.Bus, feature, err))
				}
				continue
			}
			label := "Helligkeit"
			switch feature {
			case monitor.Contrast:
				label = "Kontrast"
			case monitor.Volume:
				label = "Monitorlautstärke"
			}
			m.sliders = append(m.sliders, &slider{label: label, model: display.Model, feature: feature, bus: display.Bus, value: value, updates: make(chan int, 1)})
		}
	}
	if value, err := laptop.Get(ctx); err == nil {
		m.sliders = append(m.sliders, &slider{label: "Laptop", laptop: true, value: value, updates: make(chan int, 1)})
	}
	if len(m.sliders) == 0 && m.Message() == "" {
		m.setMessage("Keine Regler gefunden")
	}
	return m
}

func (m *Model) Message() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.message
}

func (m *Model) setMessage(message string) {
	m.mu.Lock()
	m.message = message
	m.mu.Unlock()
	if m.changed != nil {
		m.changed()
	}
}

// Request keeps only the latest pending value for this slider.
func (m *Model) Request(index, percent int) {
	s := m.sliders[index]
	s.value = percent
	for {
		select {
		case s.updates <- percent:
			return
		default:
			// The worker may have taken the value since the send was tried.
			select {
			case <-s.updates:
			default:
			}
		}
	}
}

func (m *Model) start(ctx context.Context) {
	for _, slider := range m.sliders {
		go m.process(ctx, slider)
	}
}

func (m *Model) process(ctx context.Context, slider *slider) {
	for {
		var value int
		select {
		case <-ctx.Done():
			return
		case value = <-slider.updates:
		}
		if ctx.Err() != nil {
			return
		}
		// Requests during a write replace the one pending value in the channel.
		var err error
		if slider.laptop {
			err = m.laptop.Set(ctx, value)
		} else {
			err = m.monitors.SetFeature(ctx, slider.bus, slider.feature, value)
		}
		if err != nil {
			if slider.laptop {
				m.setMessage(fmt.Sprintf("%s: %v", slider.label, err))
			} else {
				m.setMessage(fmt.Sprintf("%s (Bus %d), %s: %v", slider.model, slider.bus, slider.label, err))
			}
		} else if m.Message() != "" {
			m.setMessage("")
		}
	}
}
