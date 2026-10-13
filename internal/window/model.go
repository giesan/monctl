// Package window provides slider state for the desktop brightness window.
package window

import (
	"context"
	"fmt"
	"sync"
	"time"

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

type slider struct {
	label   string
	value   int // UI goroutine only
	bus     int
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
	displays, err := monitors.List(ctx)
	if err != nil {
		m.setMessage(fmt.Sprintf("Externe Monitore: %v", err))
	}
	for _, display := range displays {
		value, err := monitors.Get(ctx, display.Bus)
		if err != nil {
			m.setMessage(fmt.Sprintf("%s: %v", display.Model, err))
			continue
		}
		m.sliders = append(m.sliders, &slider{label: fmt.Sprintf("%s (Bus %d)", display.Model, display.Bus), bus: display.Bus, value: value, updates: make(chan int, 1)})
	}
	if value, err := laptop.Get(ctx); err == nil {
		m.sliders = append(m.sliders, &slider{label: "Laptop", laptop: true, value: value, updates: make(chan int, 1)})
	}
	if len(m.sliders) == 0 && m.Message() == "" {
		m.setMessage("Keine Helligkeitsregler gefunden")
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
	select {
	case s.updates <- percent:
	default:
		<-s.updates
		s.updates <- percent
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
		// Avoid sending a DDC transaction for every pixel while dragging.
		timer := time.NewTimer(220 * time.Millisecond)
	debounce:
		for {
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			case value = <-slider.updates:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(220 * time.Millisecond)
			case <-timer.C:
				break debounce
			}
		}
		var err error
		if slider.laptop {
			err = m.laptop.Set(ctx, value)
		} else {
			err = m.monitors.Set(ctx, slider.bus, value)
		}
		if err != nil {
			m.setMessage(fmt.Sprintf("%s: %v", slider.label, err))
		} else if m.Message() != "" {
			m.setMessage("")
		}
	}
}
