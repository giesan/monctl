//go:build gtk

package window

import (
	"context"
	"runtime"
	"testing"

	"github.com/giesan/monctl/internal/monitor"
	"github.com/gotk3/gotk3/gtk"
)

func TestGTKPercentageLabelStaysOnOneLine(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := gtk.InitCheck(nil); err != nil {
		t.Skipf("no graphical display: %v", err)
	}
	label, err := percentageLabel(100)
	if err != nil {
		t.Fatal(err)
	}
	text, err := label.GetText()
	if err != nil || text != "100%" || !label.GetSingleLineMode() || label.GetLineWrap() || label.GetWidthChars() < 5 {
		t.Fatalf("label text=%q width=%d single=%v wrap=%v err=%v", text, label.GetWidthChars(), label.GetSingleLineMode(), label.GetLineWrap(), err)
	}
}

func TestGTKCardsDoNotWriteDuringConstruction(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := gtk.InitCheck(nil); err != nil {
		t.Skipf("no graphical display: %v", err)
	}
	provider, err := gtk.CssProviderNew()
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.LoadFromData(windowCSS); err != nil {
		t.Fatalf("card CSS: %v", err)
	}
	backend := &fakeMonitors{
		displays: []monitor.Display{{Bus: 7, Model: "LG"}},
		values: map[controlKey]int{
			{7, monitor.Brightness}: 35,
			{7, monitor.Contrast}:   60,
			{7, monitor.Volume}:     25,
		},
	}
	state := New(context.Background(), backend, &fakeLaptop{value: 50}, nil)
	content, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	if err != nil {
		t.Fatal(err)
	}
	if err := populateWindow(content, state); err != nil {
		t.Fatal(err)
	}
	if count := content.GetChildren().Length(); count != 2 || len(backend.writes) != 0 {
		t.Fatalf("cards=%d writes=%v", count, backend.writes)
	}
}

func TestGTKPreferredHeightTracksVisibleCards(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := gtk.InitCheck(nil); err != nil {
		t.Skipf("no graphical display: %v", err)
	}
	win, err := gtk.OffscreenWindowNew()
	if err != nil {
		t.Fatal(err)
	}
	defer win.Destroy()
	scroller, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	scroller.SetPropagateNaturalHeight(true)
	content, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	if err != nil {
		t.Fatal(err)
	}
	scroller.Add(content)
	win.Add(scroller)
	first, err := gtk.LabelNew("LG Ultra HD\nBrightness\nContrast\nVolume")
	if err != nil {
		t.Fatal(err)
	}
	content.PackStart(first, false, false, 0)
	win.ShowAll()
	_, oneCard := scroller.GetPreferredHeightForWidth(480)
	second, err := gtk.LabelNew("Laptop display\nBrightness")
	if err != nil {
		t.Fatal(err)
	}
	content.PackStart(second, false, false, 0)
	second.Show()
	_, twoCards := scroller.GetPreferredHeightForWidth(480)
	if twoCards <= oneCard {
		t.Fatalf("GTK natural height did not grow with content: one=%d two=%d", oneCard, twoCards)
	}
}
