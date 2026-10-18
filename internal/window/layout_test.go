package window

import (
	"reflect"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

func TestGroupSliders(t *testing.T) {
	controls := []*slider{
		{model: "LG", bus: 7, feature: monitor.Brightness},
		{model: "LG", bus: 7, feature: monitor.Contrast},
		{model: "LG", bus: 7, feature: monitor.Volume},
		{model: "PHL", bus: 8, feature: monitor.Volume},
		{laptop: true, label: "Laptop"},
	}
	want := []deviceGroup{
		{name: "LG (Bus 7)", sliders: []int{0, 1, 2}},
		{name: "PHL (Bus 8)", sliders: []int{3}},
		{name: "Laptop-Bildschirm", sliders: []int{4}},
	}
	if got := groupSliders(controls); !reflect.DeepEqual(got, want) {
		t.Fatalf("groups = %+v, want %+v", got, want)
	}
	if got := groupSliders(nil); len(got) != 0 {
		t.Fatalf("empty groups = %+v", got)
	}
}

func TestFormatPercent(t *testing.T) {
	for _, tc := range []struct {
		value int
		want  string
	}{{0, "0 %"}, {45, "45 %"}, {100, "100 %"}} {
		if got := formatPercent(tc.value); got != tc.want {
			t.Fatalf("formatPercent(%d) = %q, want %q", tc.value, got, tc.want)
		}
	}
}
