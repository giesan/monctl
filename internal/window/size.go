package window

// fitHeight uses GTK's natural widget height unless the screen's work area
// requires scrolling. GTK's minimum is still respected for the visible controls.
func fitHeight(minimum, natural, workarea int) int {
	if natural < minimum {
		natural = minimum
	}
	if workarea <= 0 {
		return natural
	}
	// Reserve space for the title bar, window decorations and desktop panels.
	limit := max(minimum, workarea-80)
	return min(natural, limit)
}
