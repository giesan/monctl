package tray

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	itemInterface = "org.kde.StatusNotifierItem"
	menuInterface = "com.canonical.dbusmenu"
	itemPath      = dbus.ObjectPath("/org/ayatana/NotificationItem/monctl")
	menuPath      = dbus.ObjectPath("/MenuBar")
	watcherPath   = dbus.ObjectPath("/StatusNotifierWatcher")
	busName       = "org.giesan.Monctl.Tray"
)

type pixmap struct {
	Width  int32
	Height int32
	Bytes  []byte
}

// Pixmaps are ARGB bytes in network order, as required by StatusNotifierItem.
func icon() pixmap {
	const size = 16
	data := make([]byte, size*size*4)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := x-7, y-7
			distance := dx*dx + dy*dy
			if distance > 13 && !(distance >= 25 && distance <= 43 && (x == 7 || y == 7 || dx*dx == dy*dy)) {
				continue
			}
			index := (y*size + x) * 4
			data[index], data[index+1], data[index+2], data[index+3] = 255, 245, 190, 40
		}
	}
	return pixmap{Width: size, Height: size, Bytes: data}
}

type tooltip struct {
	IconName string
	Pixmaps  []pixmap
	Title    string
	Body     string
}

type menuLayout struct {
	ID       int32
	Props    map[string]dbus.Variant
	Children []dbus.Variant
}

type menuProperty struct {
	ID    int32
	Props map[string]dbus.Variant
}

type menuEvent struct {
	ID        int32
	Name      string
	Data      dbus.Variant
	Timestamp uint32
}

type Server struct {
	conn       *dbus.Conn
	model      model
	mu         sync.Mutex
	actionMu   sync.Mutex
	version    uint32
	quit       chan struct{}
	once       sync.Once
	openWindow func() error
}

func Start(ctx context.Context, monitors monitors, laptop laptop) (*Server, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("connect to desktop session bus: %w", err)
	}
	s := &Server{conn: conn, model: model{monitors: monitors, laptop: laptop}, quit: make(chan struct{}), version: 1, openWindow: openWindow}
	s.model.refresh(ctx)
	if err := s.export(); err != nil {
		conn.Close()
		return nil, err
	}
	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		conn.Close()
		return nil, fmt.Errorf("tray is already running or could not claim its D-Bus name: %v", err)
	}
	call := conn.Object("org.kde.StatusNotifierWatcher", watcherPath).
		CallWithContext(ctx, "org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, string(itemPath))
	if call.Err != nil {
		conn.Close()
		return nil, fmt.Errorf("register tray icon: %w (in Cinnamon, enable the XApp Status Applet in the panel)", call.Err)
	}
	return s, nil
}

func (s *Server) export() error {
	itemProps, err := prop.Export(s.conn, itemPath, prop.Map{itemInterface: {
		"Category":            {Value: "ApplicationStatus"},
		"Id":                  {Value: "monctl"},
		"Title":               {Value: "Monitor Control"},
		"Status":              {Value: "Active"},
		"WindowId":            {Value: int32(0)},
		"IconName":            {Value: "display-brightness-symbolic"},
		"IconThemePath":       {Value: ""},
		"IconPixmap":          {Value: []pixmap{icon()}},
		"OverlayIconName":     {Value: ""},
		"OverlayIconPixmap":   {Value: []pixmap{}},
		"AttentionIconName":   {Value: ""},
		"AttentionIconPixmap": {Value: []pixmap{}},
		"ToolTip":             {Value: tooltip{Title: "Monitor Control", Pixmaps: []pixmap{icon()}}},
		"Menu":                {Value: menuPath},
		"ItemIsMenu":          {Value: true},
	}})
	if err != nil {
		return err
	}
	menuProps, err := prop.Export(s.conn, menuPath, prop.Map{menuInterface: {
		"Version":       {Value: uint32(3)},
		"TextDirection": {Value: "ltr"},
		"Status":        {Value: "normal"},
		"IconThemePath": {Value: []string{}},
	}})
	if err != nil {
		return err
	}
	if err := s.conn.Export((*itemMethods)(s), itemPath, itemInterface); err != nil {
		return err
	}
	if err := s.conn.Export((*menuMethods)(s), menuPath, menuInterface); err != nil {
		return err
	}
	for _, info := range []struct {
		path  dbus.ObjectPath
		iface introspect.Interface
	}{
		{itemPath, introspect.Interface{Name: itemInterface, Methods: introspect.Methods((*itemMethods)(s)), Properties: itemProps.Introspection(itemInterface)}},
		{menuPath, introspect.Interface{Name: menuInterface, Methods: introspect.Methods((*menuMethods)(s)), Properties: menuProps.Introspection(menuInterface)}},
	} {
		node := &introspect.Node{Name: string(info.path), Interfaces: []introspect.Interface{info.iface, introspect.IntrospectData, prop.IntrospectData}}
		if err := s.conn.Export(introspect.NewIntrospectable(node), info.path, "org.freedesktop.DBus.Introspectable"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Wait(ctx context.Context) error {
	select {
	case <-s.quit:
	case <-ctx.Done():
	}
	return s.conn.Close()
}

type itemMethods Server

func (*itemMethods) Activate(int32, int32) *dbus.Error          { return nil }
func (*itemMethods) SecondaryActivate(int32, int32) *dbus.Error { return nil }
func (*itemMethods) ContextMenu(int32, int32) *dbus.Error       { return nil }
func (*itemMethods) Scroll(int32, string) *dbus.Error           { return nil }

type menuMethods Server

func (s *Server) properties(item entry, names []string) map[string]dbus.Variant {
	all := map[string]dbus.Variant{
		"label":   dbus.MakeVariant(item.label),
		"enabled": dbus.MakeVariant(item.enabled),
	}
	if len(names) == 0 {
		return all
	}
	selected := make(map[string]dbus.Variant)
	for _, name := range names {
		if value, ok := all[name]; ok {
			selected[name] = value
		}
	}
	return selected
}

func (m *menuMethods) GetLayout(parentID, depth int32, names []string) (uint32, menuLayout, *dbus.Error) {
	s := (*Server)(m)
	s.mu.Lock()
	defer s.mu.Unlock()
	root := menuLayout{ID: 0, Props: map[string]dbus.Variant{"children-display": dbus.MakeVariant("submenu")}, Children: []dbus.Variant{}}
	if parentID != 0 {
		for _, item := range s.model.entries {
			if item.id == parentID {
				return s.version, menuLayout{ID: item.id, Props: s.properties(item, names), Children: []dbus.Variant{}}, nil
			}
		}
		return s.version, root, dbus.NewError("com.canonical.dbusmenu.Error.InvalidItem", nil)
	}
	if depth != 0 {
		for _, item := range s.model.entries {
			root.Children = append(root.Children, dbus.MakeVariant(menuLayout{ID: item.id, Props: s.properties(item, names), Children: []dbus.Variant{}}))
		}
	}
	return s.version, root, nil
}

func (m *menuMethods) GetGroupProperties(ids []int32, names []string) ([]menuProperty, *dbus.Error) {
	s := (*Server)(m)
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []menuProperty
	for _, item := range s.model.entries {
		if len(ids) == 0 || contains(ids, item.id) {
			result = append(result, menuProperty{ID: item.id, Props: s.properties(item, names)})
		}
	}
	return result, nil
}

func (m *menuMethods) GetProperty(id int32, name string) (dbus.Variant, *dbus.Error) {
	s := (*Server)(m)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.model.entries {
		if item.id == id {
			if value, ok := s.properties(item, nil)[name]; ok {
				return value, nil
			}
		}
	}
	return dbus.Variant{}, dbus.NewError("com.canonical.dbusmenu.Error.InvalidProperty", nil)
}

func (m *menuMethods) AboutToShow(int32) (bool, *dbus.Error) {
	s := (*Server)(m)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.refresh(ctx)
	return true, nil
}

func (m *menuMethods) AboutToShowGroup(ids []int32) ([]int32, []int32, *dbus.Error) {
	_, _ = m.AboutToShow(0)
	return ids, []int32{}, nil
}

func (m *menuMethods) Event(id int32, name string, data dbus.Variant, timestamp uint32) *dbus.Error {
	if name != "clicked" {
		return nil
	}
	s := (*Server)(m)
	s.mu.Lock()
	act, ok := s.model.actions[id]
	s.mu.Unlock()
	if !ok {
		return nil
	}
	if act.kind == "quit" {
		s.once.Do(func() { close(s.quit) })
		return nil
	}
	if act.kind == "window" {
		go func() {
			if err := s.openWindow(); err != nil {
				log.Printf("monctl-tray: %v", err)
			}
		}()
		return nil
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.actionMu.Lock()
		defer s.actionMu.Unlock()
		if err := s.model.apply(ctx, act); err != nil {
			log.Printf("monctl-tray: %v", err)
		}
		s.refresh(ctx)
	}()
	return nil
}

func (m *menuMethods) EventGroup(events []menuEvent) ([]int32, *dbus.Error) {
	for _, event := range events {
		_ = m.Event(event.ID, event.Name, event.Data, event.Timestamp)
	}
	return []int32{}, nil
}

func (s *Server) refresh(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.model.refresh(ctx)
	s.version++
	if err := s.conn.Emit(menuPath, menuInterface+".LayoutUpdated", s.version, int32(0)); err != nil && !errors.Is(err, dbus.ErrClosed) {
		log.Printf("monctl-tray: refresh menu: %v", err)
	}
}

func contains(ids []int32, id int32) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}
