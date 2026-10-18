package window

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type controlKey struct {
	bus     int
	feature monitor.Feature
}

type controlWrite struct {
	controlKey
	value int
}

type fakeMonitors struct {
	displays        []monitor.Display
	values          map[controlKey]int
	maxima          map[controlKey]int
	errors          map[controlKey]error
	setError        error
	mu              sync.Mutex
	writes          []controlWrite
	writtenMaximums []int
	reads           []controlKey
}

func (f *fakeMonitors) Detect(context.Context) ([]monitor.Display, error) {
	return f.displays, nil
}
func (f *fakeMonitors) ReadFeature(_ context.Context, bus int, feature monitor.Feature) (monitor.FeatureValue, error) {
	key := controlKey{bus, feature}
	f.reads = append(f.reads, key)
	maximum := f.maxima[key]
	if maximum == 0 {
		maximum = 100
	}
	return monitor.FeatureValue{Percent: f.values[key], Maximum: maximum}, f.errors[key]
}
func (f *fakeMonitors) SetFeatureWithMaximum(_ context.Context, bus int, feature monitor.Feature, maximum, value int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, controlWrite{controlKey{bus, feature}, value})
	f.writtenMaximums = append(f.writtenMaximums, maximum)
	return f.setError
}

type fakeLaptop struct {
	mu     sync.Mutex
	value  int
	writes chan int
}

func (f *fakeLaptop) Get(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.value, nil
}
func (f *fakeLaptop) Set(_ context.Context, value int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.value = value
	if f.writes != nil {
		f.writes <- value
	}
	return nil
}

func TestWindowShowsSupportedFeaturesPerMonitor(t *testing.T) {
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}, {Bus: 8, Model: "AOC"}},
		values: map[controlKey]int{
			{7, monitor.Brightness}: 30, {7, monitor.Contrast}: 45,
			{8, monitor.Volume}: 65,
		},
		errors: map[controlKey]error{
			{7, monitor.Volume}:     monitor.ErrUnsupported,
			{8, monitor.Brightness}: monitor.ErrUnsupported,
			{8, monitor.Contrast}:   monitor.ErrUnsupported,
		},
	}
	m := New(context.Background(), monitors, &fakeLaptop{value: 70}, nil)
	if len(monitors.reads) != 6 || len(m.sliders) != 4 || m.Message() != "" {
		t.Fatalf("reads=%v sliders=%+v message=%q", monitors.reads, m.sliders, m.Message())
	}
	for i, want := range []struct {
		label, model string
		bus, value   int
		feature      monitor.Feature
	}{
		{"Helligkeit", "PHL", 7, 30, monitor.Brightness},
		{"Kontrast", "PHL", 7, 45, monitor.Contrast},
		{"Monitorlautstärke", "AOC", 8, 65, monitor.Volume},
		{"Laptop", "", 0, 70, 0},
	} {
		s := m.sliders[i]
		if s.label != want.label || s.model != want.model || s.bus != want.bus || s.value != want.value || s.feature != want.feature {
			t.Fatalf("slider %d = %+v, want %+v", i, s, want)
		}
	}
}

func TestWindowWritesWithStartupMaximumWithoutRereading(t *testing.T) {
	key := controlKey{7, monitor.Contrast}
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		values:   map[controlKey]int{key: 25},
		maxima:   map[controlKey]int{key: 80},
		errors: map[controlKey]error{
			{7, monitor.Brightness}: monitor.ErrUnsupported,
			{7, monitor.Volume}:     monitor.ErrUnsupported,
		},
	}
	m := New(context.Background(), monitors, unavailableLaptop{}, nil)
	if len(m.sliders) != 1 || m.sliders[0].maximum != 80 || len(monitors.reads) != 3 {
		t.Fatalf("sliders=%+v reads=%v", m.sliders, monitors.reads)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(0, 50)
	deadline := time.After(2 * time.Second)
	for {
		monitors.mu.Lock()
		written := append([]controlWrite(nil), monitors.writes...)
		maxima := append([]int(nil), monitors.writtenMaximums...)
		monitors.mu.Unlock()
		if len(written) > 0 {
			if len(written) != 1 || written[0] != (controlWrite{key, 50}) || len(maxima) != 1 || maxima[0] != 80 || len(monitors.reads) != 3 {
				t.Fatalf("writes=%v maxima=%v reads=%v", written, maxima, monitors.reads)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("monitor write did not happen")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestWindowKeepsWorkingAfterFeatureReadError(t *testing.T) {
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		values:   map[controlKey]int{{7, monitor.Brightness}: 30, {7, monitor.Volume}: 60},
		errors:   map[controlKey]error{{7, monitor.Contrast}: errors.New("read failed")},
	}
	m := New(context.Background(), monitors, &fakeLaptop{value: 70}, nil)
	if len(m.sliders) != 3 || !strings.Contains(m.Message(), "PHL (Bus 7), contrast: read failed") {
		t.Fatalf("sliders=%+v message=%q", m.sliders, m.Message())
	}
}

func TestWindowWithoutSupportedControls(t *testing.T) {
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		errors: map[controlKey]error{
			{7, monitor.Brightness}: monitor.ErrUnsupported,
			{7, monitor.Contrast}:   monitor.ErrUnsupported,
			{7, monitor.Volume}:     monitor.ErrUnsupported,
		},
	}
	m := New(context.Background(), monitors, unavailableLaptop{}, nil)
	if len(m.sliders) != 0 || m.Message() != "Keine Regler gefunden" {
		t.Fatalf("sliders=%+v message=%q", m.sliders, m.Message())
	}
}

type unavailableLaptop struct{}

func (unavailableLaptop) Get(context.Context) (int, error) { return 0, errors.New("no backlight") }
func (unavailableLaptop) Set(context.Context, int) error   { return errors.New("no backlight") }

type blockingMonitors struct {
	*fakeMonitors
	started chan controlWrite
	release chan struct{}
}

func (b *blockingMonitors) SetFeatureWithMaximum(ctx context.Context, bus int, feature monitor.Feature, maximum, value int) error {
	write := controlWrite{controlKey{bus, feature}, value}
	b.started <- write
	if value == 40 {
		select {
		case <-b.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return b.fakeMonitors.SetFeatureWithMaximum(ctx, bus, feature, maximum, value)
}

func waitWrite(t *testing.T, writes <-chan controlWrite) controlWrite {
	t.Helper()
	select {
	case write := <-writes:
		return write
	case <-time.After(200 * time.Millisecond):
		t.Fatal("write did not start immediately")
		return controlWrite{}
	}
}

func TestSliderWritesImmediatelyAndKeepsLatestPendingValue(t *testing.T) {
	monitors := &blockingMonitors{fakeMonitors: &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		values: map[controlKey]int{
			{7, monitor.Brightness}: 30, {7, monitor.Contrast}: 40, {7, monitor.Volume}: 60,
		},
	}, started: make(chan controlWrite, 4), release: make(chan struct{})}
	m := New(context.Background(), monitors, &fakeLaptop{value: 70}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(2, 40)
	if got := waitWrite(t, monitors.started); got != (controlWrite{controlKey{7, monitor.Volume}, 40}) {
		t.Fatalf("first write = %v", got)
	}
	m.Request(2, 45)
	m.Request(2, 50)
	m.Request(2, 75)
	close(monitors.release)
	if got := waitWrite(t, monitors.started); got != (controlWrite{controlKey{7, monitor.Volume}, 75}) {
		t.Fatalf("second write = %v, want volume 75", got)
	}
	select {
	case unexpected := <-monitors.started:
		t.Fatalf("obsolete value written: %v", unexpected)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestLaptopSliderWritesImmediately(t *testing.T) {
	laptop := &fakeLaptop{value: 70, writes: make(chan int, 1)}
	m := New(context.Background(), &fakeMonitors{}, laptop, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(0, 80)
	select {
	case value := <-laptop.writes:
		if value != 80 {
			t.Fatalf("laptop write = %d", value)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("laptop write did not start immediately")
	}
}

func TestCancelDropsPendingSliderValue(t *testing.T) {
	monitors := &blockingMonitors{fakeMonitors: &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		values:   map[controlKey]int{{7, monitor.Brightness}: 30},
		errors:   map[controlKey]error{{7, monitor.Contrast}: monitor.ErrUnsupported, {7, monitor.Volume}: monitor.ErrUnsupported},
	}, started: make(chan controlWrite, 2), release: make(chan struct{})}
	m := New(context.Background(), monitors, unavailableLaptop{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	m.start(ctx)
	m.Request(0, 40)
	waitWrite(t, monitors.started)
	m.Request(0, 75)
	cancel()
	select {
	case unexpected := <-monitors.started:
		t.Fatalf("write started after cancellation: %v", unexpected)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestWriteErrorNamesMonitorAndControl(t *testing.T) {
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}},
		values:   map[controlKey]int{{7, monitor.Brightness}: 30},
		errors:   map[controlKey]error{{7, monitor.Contrast}: monitor.ErrUnsupported, {7, monitor.Volume}: monitor.ErrUnsupported},
		setError: errors.New("write failed"),
	}
	m := New(context.Background(), monitors, &fakeLaptop{value: 70}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(0, 40)
	deadline := time.After(2 * time.Second)
	for !strings.Contains(m.Message(), "PHL (Bus 7), Helligkeit: write failed") {
		select {
		case <-deadline:
			t.Fatalf("write error = %q", m.Message())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestWriteTargetsSelectedMonitor(t *testing.T) {
	monitors := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "PHL"}, {Bus: 8, Model: "AOC"}},
		values: map[controlKey]int{
			{7, monitor.Brightness}: 30, {8, monitor.Contrast}: 40,
		},
		errors: map[controlKey]error{
			{7, monitor.Contrast}: monitor.ErrUnsupported, {7, monitor.Volume}: monitor.ErrUnsupported,
			{8, monitor.Brightness}: monitor.ErrUnsupported, {8, monitor.Volume}: monitor.ErrUnsupported,
		},
	}
	m := New(context.Background(), monitors, unavailableLaptop{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.start(ctx)
	m.Request(1, 65)
	deadline := time.After(2 * time.Second)
	for {
		monitors.mu.Lock()
		written := append([]controlWrite(nil), monitors.writes...)
		monitors.mu.Unlock()
		if len(written) > 0 {
			if len(written) != 1 || written[0] != (controlWrite{controlKey{8, monitor.Contrast}, 65}) {
				t.Fatalf("writes = %v", written)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("monitor write did not happen")
		case <-time.After(10 * time.Millisecond):
		}
	}
}
