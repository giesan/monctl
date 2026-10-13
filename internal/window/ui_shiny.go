//go:build nucular_shiny

package window

import (
	"context"
	"fmt"
	"image"
	"os"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/style"
)

// Run opens an X11 window using Nucular's CGO-free Shiny backend.
func Run(monitors monitors, laptop laptop) error {
	if os.Getenv("DISPLAY") == "" {
		return fmt.Errorf("no X11 display available; an X11 or XWayland session is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state := New(ctx, monitors, laptop, nil)
	height := min(680, 110+len(state.sliders)*78)
	wnd := nucular.NewMasterWindowSize(0, "Monitor Control", image.Pt(480, height), func(w *nucular.Window) {
		for index, slider := range state.sliders {
			w.Row(30).Dynamic(1)
			w.Label(fmt.Sprintf("%s — %d%%", slider.label, slider.value), label.Align("LC"))
			w.Row(38).Dynamic(1)
			value := slider.value
			if w.SliderInt(0, &value, 100, 1) {
				state.Request(index, value)
			}
		}
		if message := state.Message(); message != "" {
			w.Row(42).Dynamic(1)
			w.Label(message, label.Align("LC"))
		}
	})
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 1.0))
	state.changed = wnd.Changed
	wnd.OnClose(cancel)
	state.start(ctx)
	wnd.Main()
	return nil
}
