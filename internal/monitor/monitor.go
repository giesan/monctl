// Package monitor provides brightness operations independent of the hardware backend.
package monitor

import (
	"context"
	"fmt"
	"strings"
)

type Display struct {
	Bus   int
	Model string
}

type Brightness struct {
	Current int
	Maximum int
}

type Backend interface {
	Detect(context.Context) ([]Display, error)
	Read(context.Context, int) (Brightness, error)
	Write(context.Context, int, int) error
}

type Service struct {
	Backend Backend
}

func (s Service) List(ctx context.Context) ([]Display, error) {
	displays, err := s.Backend.Detect(ctx)
	if err != nil {
		return nil, err
	}
	var supported []Display
	for _, display := range displays {
		_, err := s.Backend.Read(ctx, display.Bus)
		if err != nil {
			// A display without brightness control is not a brightness target.
			if IsUnsupported(err) {
				continue
			}
			return nil, fmt.Errorf("probe bus %d: %w", display.Bus, err)
		}
		supported = append(supported, display)
	}
	return supported, nil
}

// Single selects a brightness-capable monitor only when the choice is unambiguous.
func (s Service) Single(ctx context.Context) (Display, error) {
	displays, err := s.List(ctx)
	if err != nil {
		return Display{}, err
	}
	switch len(displays) {
	case 0:
		return Display{}, fmt.Errorf("no DDC/CI brightness-capable monitors found; run monctl list")
	case 1:
		return displays[0], nil
	default:
		buses := make([]string, 0, len(displays))
		for _, display := range displays {
			buses = append(buses, fmt.Sprint(display.Bus))
		}
		return Display{}, fmt.Errorf("multiple monitors found (buses %s); specify a bus from monctl list", strings.Join(buses, ", "))
	}
}

func (s Service) Get(ctx context.Context, bus int) (int, error) {
	brightness, err := s.Backend.Read(ctx, bus)
	if err != nil {
		return 0, err
	}
	return percent(brightness)
}

func (s Service) Set(ctx context.Context, bus, requested int) error {
	if requested < 0 || requested > 100 {
		return fmt.Errorf("brightness must be between 0 and 100 percent")
	}
	brightness, err := s.Backend.Read(ctx, bus)
	if err != nil {
		return err
	}
	if brightness.Maximum <= 0 {
		return fmt.Errorf("monitor reported invalid maximum brightness %d", brightness.Maximum)
	}
	value := (requested*brightness.Maximum + 50) / 100
	return s.Backend.Write(ctx, bus, value)
}

func percent(b Brightness) (int, error) {
	if b.Maximum <= 0 || b.Current < 0 || b.Current > b.Maximum {
		return 0, fmt.Errorf("monitor reported invalid brightness %d/%d", b.Current, b.Maximum)
	}
	return (b.Current*100 + b.Maximum/2) / b.Maximum, nil
}
