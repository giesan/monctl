package window

import "fmt"

type deviceGroup struct {
	name    string
	sliders []int
}

func groupSliders(sliders []*slider) []deviceGroup {
	var groups []deviceGroup
	for index, control := range sliders {
		name := "Laptop-Bildschirm"
		if !control.laptop {
			name = fmt.Sprintf("%s (Bus %d)", control.model, control.bus)
		}
		if len(groups) == 0 || groups[len(groups)-1].name != name || control.laptop != sliders[groups[len(groups)-1].sliders[0]].laptop {
			groups = append(groups, deviceGroup{name: name})
		}
		groups[len(groups)-1].sliders = append(groups[len(groups)-1].sliders, index)
	}
	return groups
}

func formatPercent(value int) string {
	return fmt.Sprintf("%d %%", value)
}
