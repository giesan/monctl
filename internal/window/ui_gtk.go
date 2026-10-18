//go:build gtk

package window

import (
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const windowCSS = `
.monctl-card {
  background-color: @theme_base_color;
  color: @theme_text_color;
  border: 1px solid alpha(@theme_fg_color, 0.16);
  border-radius: 13px;
  padding: 14px;
}
.monctl-title { font-weight: bold; }
.monctl-value { font-feature-settings: "tnum"; }
`

// Run opens a Cinnamon-themed GTK 3 window on X11 or Wayland.
func Run(monitors monitors, laptop laptop) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	glib.SetPrgname("monctl")
	glib.SetApplicationName("Monitor Control")
	if err := gtk.InitCheck(nil); err != nil {
		return fmt.Errorf("no graphical GTK display available: %w", err)
	}
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	if err := provider.LoadFromData(windowCSS); err != nil {
		return fmt.Errorf("load window style: %w", err)
	}
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return fmt.Errorf("find GTK screen: %w", err)
	}
	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}
	win.SetTitle("Monitor Control")
	if err := setWindowIcons(win); err != nil {
		return err
	}
	win.SetDefaultSize(480, 220)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	win.Connect("destroy", func() {
		cancel()
		gtk.MainQuit()
	})

	outer, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return err
	}
	scroller, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return err
	}
	scroller.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	scroller.SetPropagateNaturalHeight(true)
	content, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	if err != nil {
		return err
	}
	content.SetMarginTop(16)
	content.SetMarginBottom(16)
	content.SetMarginStart(16)
	content.SetMarginEnd(16)
	scroller.Add(content)
	outer.PackStart(scroller, true, true, 0)
	status, err := gtk.LabelNew("")
	if err != nil {
		return err
	}
	status.SetLineWrap(true)
	status.SetXAlign(0)
	status.SetMarginStart(16)
	status.SetMarginEnd(16)
	status.SetMarginBottom(12)
	status.SetNoShowAll(true)
	outer.PackStart(status, false, false, 0)
	win.Add(outer)

	loading, err := gtk.LabelNew("Lade Monitore …")
	if err != nil {
		return err
	}
	loading.SetMarginTop(24)
	content.PackStart(loading, false, false, 0)
	win.ShowAll()
	lastAutoHeight := 0
	resizeToContent := func() {
		width, currentHeight := win.GetSize()
		minimumWidth, _ := outer.GetPreferredWidth()
		width = max(width, minimumWidth)
		minimumHeight, naturalHeight := outer.GetPreferredHeightForWidth(width)
		target := fitHeight(minimumHeight, naturalHeight, monitorWorkareaHeight(win))
		// Preserve a height deliberately chosen by the user after initial layout.
		if lastAutoHeight != 0 && (currentHeight > lastAutoHeight+4 || currentHeight < lastAutoHeight-4) && target <= currentHeight {
			return
		}
		if target != currentHeight {
			win.Resize(width, target)
		}
		lastAutoHeight = target
	}

	// Discovery can take seconds on some DDC devices. Never run it on the GTK thread.
	go func() {
		state := New(ctx, monitors, laptop, nil)
		glib.IdleAdd(func() {
			if ctx.Err() != nil {
				return
			}
			loading.Destroy()
			if err := populateWindow(content, state); err != nil {
				status.SetText(fmt.Sprintf("Fenster: %v", err))
				status.Show()
				return
			}
			showMessage := func() {
				if ctx.Err() != nil {
					return
				}
				if message := state.Message(); message != "" {
					status.SetText(message)
					status.Show()
				} else {
					status.Hide()
				}
				resizeToContent()
			}
			state.changed = func() {
				if ctx.Err() == nil {
					glib.IdleAdd(showMessage)
				}
			}
			content.ShowAll()
			showMessage()
			state.start(ctx)
		})
	}()
	gtk.Main()
	return nil
}

func monitorWorkareaHeight(win *gtk.Window) int {
	display, err := win.GetDisplay()
	if err != nil {
		return 0
	}
	if gdkWindow, err := win.GetWindow(); err == nil {
		if monitor, err := display.GetMonitorAtWindow(gdkWindow); err == nil {
			return monitor.GetWorkarea().GetHeight()
		}
	}
	if monitor, err := display.GetPrimaryMonitor(); err == nil {
		return monitor.GetWorkarea().GetHeight()
	}
	return 0
}

func populateWindow(content *gtk.Box, state *Model) error {
	for _, group := range groupSliders(state.sliders) {
		card, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
		if err != nil {
			return err
		}
		style, err := card.GetStyleContext()
		if err != nil {
			return err
		}
		style.AddClass("monctl-card")
		heading, err := gtk.LabelNew(group.name)
		if err != nil {
			return err
		}
		heading.SetXAlign(0)
		heading.SetLineWrap(false)
		headingStyle, err := heading.GetStyleContext()
		if err != nil {
			return err
		}
		headingStyle.AddClass("monctl-title")
		card.PackStart(heading, false, false, 0)
		for _, index := range group.sliders {
			row, err := sliderRow(state, index)
			if err != nil {
				return err
			}
			card.PackStart(row, false, false, 0)
		}
		content.PackStart(card, false, false, 0)
	}
	return nil
}

func sliderRow(state *Model, index int) (*gtk.Box, error) {
	control := state.sliders[index]
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		return nil, err
	}
	icon, err := controlIcon(control)
	if err != nil {
		return nil, err
	}
	row.PackStart(icon, false, false, 0)
	caption := control.label
	if control.laptop {
		caption = "Helligkeit"
	}
	name, err := gtk.LabelNewWithMnemonic("_" + caption)
	if err != nil {
		return nil, err
	}
	name.SetXAlign(0)
	name.SetSizeRequest(110, -1)
	row.PackStart(name, false, false, 0)
	scale, err := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 100, 1)
	if err != nil {
		return nil, err
	}
	scale.SetDrawValue(false)
	scale.SetHExpand(true)
	scale.SetSizeRequest(140, -1)
	scale.SetTooltipText(caption)
	scale.SetValue(float64(control.value)) // Connect only after setting the initial value.
	name.SetMnemonicWidget(scale)
	row.PackStart(scale, true, true, 0)
	value, err := percentageLabel(control.value)
	if err != nil {
		return nil, err
	}
	row.PackStart(value, false, false, 0)
	scale.Connect("value-changed", func() {
		percent := int(math.Round(scale.GetValue()))
		value.SetText(formatPercent(percent))
		state.Request(index, percent)
	})
	return row, nil
}

func percentageLabel(percent int) (*gtk.Label, error) {
	value, err := gtk.LabelNew(formatPercent(percent))
	if err != nil {
		return nil, err
	}
	value.SetWidthChars(5)
	value.SetSingleLineMode(true)
	value.SetLineWrap(false)
	value.SetXAlign(1)
	style, err := value.GetStyleContext()
	if err != nil {
		return nil, err
	}
	style.AddClass("monctl-value")
	return value, nil
}

func controlIcon(control *slider) (gtk.IWidget, error) {
	iconName, fallback := "display-brightness-symbolic", "☀"
	switch control.feature {
	case monitor.Contrast:
		iconName, fallback = "view-contrast-symbolic", "◐"
	case monitor.Volume:
		iconName, fallback = "audio-volume-medium-symbolic", "♪"
	}
	theme, err := gtk.IconThemeGetDefault()
	if err == nil && theme.HasIcon(iconName) {
		image, err := gtk.ImageNewFromIconName(iconName, gtk.ICON_SIZE_MENU)
		if err == nil {
			image.SetSizeRequest(24, -1)
			return image, nil
		}
	}
	label, err := gtk.LabelNew(fallback)
	if err == nil {
		label.SetSizeRequest(24, -1)
	}
	return label, err
}
