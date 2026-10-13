package ddc

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/giesan/monitorcontrol/internal/monitor"
)

type response struct {
	out string
	err error
}

type fakeRunner struct {
	responses []response
	commands  [][]string
}

func (f *fakeRunner) Run(_ context.Context, args ...string) (string, error) {
	f.commands = append(f.commands, args)
	answer := f.responses[0]
	f.responses = f.responses[1:]
	return answer.out, answer.err
}

func TestDetectBriefAndReadTerse(t *testing.T) {
	f := &fakeRunner{responses: []response{
		{out: "Display 1\n   I2C bus: /dev/i2c-7\n   Monitor: ACR:Acer X243W:123\n\nDisplay 2\n   I2C bus: /dev/i2c-12\n   Monitor: DEL:DELL U3011:456\n"},
		{out: "VCP 10 C 30 80\n"},
	}}
	b := &Backend{run: f}
	displays, err := b.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []monitor.Display{{Bus: 7, Model: "Acer X243W"}, {Bus: 12, Model: "DELL U3011"}}
	if !reflect.DeepEqual(displays, want) {
		t.Fatalf("displays = %#v, want %#v", displays, want)
	}
	brightness, err := b.Read(context.Background(), 12)
	if err != nil || brightness != (monitor.Brightness{Current: 30, Maximum: 80}) {
		t.Fatalf("brightness = %+v, err = %v", brightness, err)
	}
	if !reflect.DeepEqual(f.commands[1], []string{"getvcp", "10", "--bus", "12", "--terse"}) {
		t.Fatalf("read command = %v", f.commands[1])
	}
}

func TestReadUnsupported(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "VCP 10 ERR\n"}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7)
	if !errors.Is(err, monitor.ErrUnsupported) {
		t.Fatalf("read error = %v", err)
	}
}

func TestPermissionError(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "Permission denied: /dev/i2c-7", err: errors.New("exit status 1")}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7)
	if err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("read error = %v", err)
	}
}

func TestMissingI2CDevices(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "No /dev/i2c devices exist.\nddcutil requires module i2c-dev.", err: errors.New("exit status 1")}}}
	_, err := (&Backend{run: f}).Detect(context.Background())
	if err == nil || !strings.Contains(err.Error(), "modprobe i2c-dev") {
		t.Fatalf("detect error = %v", err)
	}
}

func TestMalformedReadIsError(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "something unexpected"}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7)
	if err == nil {
		t.Fatal("expected malformed response error")
	}
}
