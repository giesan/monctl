//go:build gtk

package window

import (
	_ "embed"
	"fmt"
	"runtime"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed icons/monctl-16.png
var icon16 []byte

//go:embed icons/monctl-24.png
var icon24 []byte

//go:embed icons/monctl-48.png
var icon48 []byte

//go:embed icons/monctl-128.png
var icon128 []byte

func setWindowIcons(win *gtk.Window) error {
	list := &glib.List{}
	var images []*gdk.Pixbuf
	for _, data := range [][]byte{icon16, icon24, icon48, icon128} {
		image, err := gdk.PixbufNewFromBytesOnly(data)
		if err != nil {
			list.Free()
			return fmt.Errorf("load embedded application icon: %w", err)
		}
		images = append(images, image)
		list = list.Append(image.Native())
	}
	gtk.WindowSetDefaultIconList(list)
	win.SetIconList(list)
	list.Free()
	runtime.KeepAlive(images)
	return nil
}
