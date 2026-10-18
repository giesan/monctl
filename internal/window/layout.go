package window

import "fmt"

type deviceGroup struct {
	name    string
	sliders []int
}

func groupSliders(sliders []*slider) []deviceGroup {
	var groups []deviceGroup
	for index, control := range sliders {
		name := "Laptop display"
		if !control.laptop {
			name = control.model
		}
		newGroup := len(groups) == 0
		if !newGroup {
			previous := sliders[groups[len(groups)-1].sliders[0]]
			newGroup = control.laptop != previous.laptop || (!control.laptop && (control.bus != previous.bus || control.model != previous.model))
		}
		if newGroup {
			groups = append(groups, deviceGroup{name: name})
		}
		groups[len(groups)-1].sliders = append(groups[len(groups)-1].sliders, index)
	}
	return groups
}

func formatPercent(value int) string {
	return fmt.Sprintf("%d%%", value)
}
