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

// Feature is a DDC/CI VCP code for a continuous monitor control.
type Feature byte

const (
	Brightness Feature = 0x10
	Contrast   Feature = 0x12
	Volume     Feature = 0x62
)

func (f Feature) String() string {
	switch f {
	case Brightness:
		return "brightness"
	case Contrast:
		return "contrast"
	case Volume:
		return "monitor volume"
	default:
		return fmt.Sprintf("VCP %02X", byte(f))
	}
}

type Level struct {
	Current int
	Maximum int
}

type Backend interface {
	Detect(context.Context) ([]Display, error)
	Read(context.Context, int, Feature) (Level, error)
	Write(context.Context, int, Feature, int) error
}

type Service struct {
	Backend Backend
}

func (s Service) List(ctx context.Context) ([]Display, error) {
	return s.ListFeature(ctx, Brightness)
}

func (s Service) ListFeature(ctx context.Context, feature Feature) ([]Display, error) {
	displays, err := s.Backend.Detect(ctx)
	if err != nil {
		return nil, err
	}
	var supported []Display
	for _, display := range displays {
		_, err := s.Backend.Read(ctx, display.Bus, feature)
		if err != nil {
			// Hide displays that do not support the selected control.
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
	return s.SingleFeature(ctx, Brightness)
}

func (s Service) SingleFeature(ctx context.Context, feature Feature) (Display, error) {
	displays, err := s.ListFeature(ctx, feature)
	if err != nil {
		return Display{}, err
	}
	switch len(displays) {
	case 0:
		return Display{}, fmt.Errorf("no DDC/CI monitors with %s control found; check the monitor menu and run monctl list", feature)
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
	return s.GetFeature(ctx, bus, Brightness)
}

func (s Service) GetFeature(ctx context.Context, bus int, feature Feature) (int, error) {
	level, err := s.Backend.Read(ctx, bus, feature)
	if err != nil {
		return 0, err
	}
	return percent(level)
}

func (s Service) Set(ctx context.Context, bus, requested int) error {
	return s.SetFeature(ctx, bus, Brightness, requested)
}

func (s Service) SetFeature(ctx context.Context, bus int, feature Feature, requested int) error {
	if requested < 0 || requested > 100 {
		return fmt.Errorf("%s must be between 0 and 100 percent", feature)
	}
	level, err := s.Backend.Read(ctx, bus, feature)
	if err != nil {
		return err
	}
	if level.Maximum <= 0 {
		return fmt.Errorf("monitor reported invalid maximum %s %d", feature, level.Maximum)
	}
	value := (requested*level.Maximum + 50) / 100
	return s.Backend.Write(ctx, bus, feature, value)
}

func percent(b Level) (int, error) {
	if b.Maximum <= 0 || b.Current < 0 || b.Current > b.Maximum {
		return 0, fmt.Errorf("monitor reported invalid brightness %d/%d", b.Current, b.Maximum)
	}
	return (b.Current*100 + b.Maximum/2) / b.Maximum, nil
}
