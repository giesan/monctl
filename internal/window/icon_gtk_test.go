//go:build gtk

package window

import (
	"bytes"
	"image/png"
	"runtime"
	"testing"

	"github.com/gotk3/gotk3/gtk"
)

func TestEmbeddedIconsAreTransparentPNGs(t *testing.T) {
	for _, tc := range []struct {
		size int
		data []byte
	}{{16, icon16}, {24, icon24}, {48, icon48}, {128, icon128}} {
		img, err := png.Decode(bytes.NewReader(tc.data))
		if err != nil {
			t.Fatalf("%d px icon: %v", tc.size, err)
		}
		if bounds := img.Bounds(); bounds.Dx() != tc.size || bounds.Dy() != tc.size {
			t.Fatalf("%d px icon bounds = %v", tc.size, bounds)
		}
		if _, _, _, alpha := img.At(0, 0).RGBA(); alpha != 0 {
			t.Fatalf("%d px icon background is not transparent", tc.size)
		}
	}
}

func TestGTKWindowUsesEmbeddedIcons(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := gtk.InitCheck(nil); err != nil {
		t.Skipf("no graphical display: %v", err)
	}
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		t.Fatal(err)
	}
	defer win.Destroy()
	if err := setWindowIcons(win); err != nil {
		t.Fatal(err)
	}
	if icons := win.GetIconList(); icons == nil || icons.Length() != 4 {
		t.Fatalf("GTK window icons = %v, want four resolutions", icons)
	}
}
