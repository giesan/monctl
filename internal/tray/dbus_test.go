package tray

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/giesan/monitorcontrol/internal/monitor"
	"github.com/godbus/dbus/v5"
)

type testWatcher struct{ item chan string }

func (w *testWatcher) RegisterStatusNotifierItem(item string) *dbus.Error {
	w.item <- item
	return nil
}

func TestTrayRegistersAndServesMenu(t *testing.T) {
	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skip("dbus-daemon not installed")
	}
	socket := filepath.Join(t.TempDir(), "bus")
	cmd := exec.Command("dbus-daemon", "--session", "--nofork", "--address=unix:path="+socket)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	for attempts := 0; attempts < 100; attempts++ {
		if _, err := os.Stat(socket); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := os.Stat(socket); err != nil {
		t.Fatalf("D-Bus socket unavailable: %v", err)
	}
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path="+socket)
	watcherConn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatal(err)
	}
	defer watcherConn.Close()
	w := &testWatcher{item: make(chan string, 1)}
	if err := watcherConn.Export(w, watcherPath, "org.kde.StatusNotifierWatcher"); err != nil {
		t.Fatal(err)
	}
	if _, err := watcherConn.RequestName("org.kde.StatusNotifierWatcher", dbus.NameFlagDoNotQueue); err != nil {
		t.Fatal(err)
	}
	monitors := &fakeMonitors{displays: []monitor.Display{{Bus: 7, Model: "PHL"}}, levels: map[int]int{7: 9}}
	server, err := Start(context.Background(), monitors, &fakeLaptop{value: 75})
	if err != nil {
		t.Fatal(err)
	}
	defer server.conn.Close()
	opened := make(chan struct{}, 1)
	server.openWindow = func() error { opened <- struct{}{}; return nil }
	if item := <-w.item; item != string(itemPath) {
		t.Fatalf("registered item = %q", item)
	}
	obj := watcherConn.Object(busName, itemPath)
	var iconName dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, itemInterface, "IconName").Store(&iconName); err != nil {
		t.Fatal(err)
	}
	if iconName.Value() != "display-brightness-symbolic" {
		t.Fatalf("icon = %v", iconName.Value())
	}
	var itemProperties map[string]dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, itemInterface).Store(&itemProperties); err != nil {
		t.Fatal(err)
	}
	if itemProperties["Menu"].Value() != menuPath {
		t.Fatalf("menu path = %v", itemProperties["Menu"].Value())
	}
	var version uint32
	var layout menuLayout
	menu := watcherConn.Object(busName, menuPath)
	if err := menu.Call(menuInterface+".GetLayout", 0, int32(0), int32(-1), []string{}).Store(&version, &layout); err != nil {
		t.Fatal(err)
	}
	if version == 0 || len(layout.Children) != 8 {
		t.Fatalf("layout version=%d children=%d", version, len(layout.Children))
	}
	last, ok := layout.Children[len(layout.Children)-1].Value().([]any)
	if !ok || len(last) != 3 {
		t.Fatalf("last menu item = %#v", layout.Children[len(layout.Children)-1].Value())
	}
	props, ok := last[1].(map[string]dbus.Variant)
	if !ok || !strings.Contains(props["label"].Value().(string), "Beenden") {
		t.Fatalf("last menu label = %#v", last[1])
	}
	if err := menu.Call(menuInterface+".Event", 0, int32(3), "clicked", dbus.MakeVariant(uint32(0)), uint32(0)).Err; err != nil {
		t.Fatal(err)
	}
	for attempts := 0; attempts < 100 && monitors.level(7) != 14; attempts++ {
		time.Sleep(10 * time.Millisecond)
	}
	if monitors.level(7) != 14 {
		t.Fatalf("external brightness = %d, want 14", monitors.level(7))
	}
	if err := menu.Call(menuInterface+".Event", 0, int32(7), "clicked", dbus.MakeVariant(uint32(0)), uint32(0)).Err; err != nil {
		t.Fatal(err)
	}
	select {
	case <-opened:
	case <-time.After(time.Second):
		t.Fatal("slider window action not received")
	}
	if err := menu.Call(menuInterface+".Event", 0, last[0].(int32), "clicked", dbus.MakeVariant(uint32(0)), uint32(0)).Err; err != nil {
		t.Fatal(err)
	}
	select {
	case <-server.quit:
	case <-time.After(time.Second):
		t.Fatal("quit action not received")
	}
	if err := server.Wait(context.Background()); err != nil && !errors.Is(err, dbus.ErrClosed) {
		t.Fatal(err)
	}
}
