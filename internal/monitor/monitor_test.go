package monitor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeBackend struct {
	displays []Display
	values   map[int]Level
	errors   map[int]error
	written  int
	feature  Feature
}

func (f *fakeBackend) Detect(context.Context) ([]Display, error) { return f.displays, nil }
func (f *fakeBackend) Read(_ context.Context, bus int, _ Feature) (Level, error) {
	return f.values[bus], f.errors[bus]
}
func (f *fakeBackend) Write(_ context.Context, _ int, feature Feature, value int) error {
	f.written = value
	f.feature = feature
	return nil
}

func TestGetSetNonStandardMaximum(t *testing.T) {
	f := &fakeBackend{values: map[int]Level{7: {Current: 30, Maximum: 80}}}
	s := Service{Backend: f}
	got, err := s.Get(context.Background(), 7)
	if err != nil || got != 38 {
		t.Fatalf("get = %d, err = %v; want 38%%", got, err)
	}
	for _, tc := range []struct{ percent, device int }{{0, 0}, {50, 40}, {100, 80}, {37, 30}} {
		if err := s.Set(context.Background(), 7, tc.percent); err != nil {
			t.Fatal(err)
		}
		if f.written != tc.device {
			t.Fatalf("set %d%% wrote %d, want %d", tc.percent, f.written, tc.device)
		}
	}
}

func TestListSkipsUnsupported(t *testing.T) {
	f := &fakeBackend{
		displays: []Display{{Bus: 7, Model: "A"}, {Bus: 8, Model: "B"}},
		values:   map[int]Level{7: {Current: 10, Maximum: 100}},
		errors:   map[int]error{8: ErrUnsupported},
	}
	got, err := (Service{Backend: f}).List(context.Background())
	if err != nil || len(got) != 1 || got[0].Bus != 7 {
		t.Fatalf("list = %v, err = %v", got, err)
	}
	f.errors[8] = errors.New("permission denied")
	if _, err := (Service{Backend: f}).List(context.Background()); err == nil {
		t.Fatal("permission error was hidden")
	}
}

func TestSingleSelectsOnlyUnambiguousDisplay(t *testing.T) {
	f := &fakeBackend{values: map[int]Level{7: {Current: 10, Maximum: 100}, 8: {Current: 20, Maximum: 100}}}
	s := Service{Backend: f}
	if _, err := s.Single(context.Background()); err == nil || !strings.Contains(err.Error(), "no DDC/CI") {
		t.Fatalf("no displays: err=%v", err)
	}
	f.displays = []Display{{Bus: 7, Model: "A"}}
	if got, err := s.Single(context.Background()); err != nil || got.Bus != 7 {
		t.Fatalf("one display: got=%v err=%v", got, err)
	}
	f.displays = append(f.displays, Display{Bus: 8, Model: "B"})
	if _, err := s.Single(context.Background()); err == nil || !strings.Contains(err.Error(), "buses 7, 8") {
		t.Fatalf("multiple displays: err=%v", err)
	}
}

func TestFeatureConversionAndWrite(t *testing.T) {
	f := &fakeBackend{values: map[int]Level{7: {Current: 25, Maximum: 80}}}
	s := Service{Backend: f}
	got, err := s.GetFeature(context.Background(), 7, Contrast)
	if err != nil || got != 31 {
		t.Fatalf("contrast=%d err=%v", got, err)
	}
	if err := s.SetFeature(context.Background(), 7, Volume, 50); err != nil || f.written != 40 || f.feature != Volume {
		t.Fatalf("volume write=%d feature=%v err=%v", f.written, f.feature, err)
	}
}
