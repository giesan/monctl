package ddc

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/giesan/monctl/internal/monitor"
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
	brightness, err := b.Read(context.Background(), 12, monitor.Brightness)
	if err != nil || brightness != (monitor.Level{Current: 30, Maximum: 80}) {
		t.Fatalf("brightness = %+v, err = %v", brightness, err)
	}
	if !reflect.DeepEqual(f.commands[1], []string{"getvcp", "10", "--bus", "12", "--terse"}) {
		t.Fatalf("read command = %v", f.commands[1])
	}
}

func TestReadUnsupported(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "VCP 10 ERR\n"}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7, monitor.Brightness)
	if !errors.Is(err, monitor.ErrUnsupported) {
		t.Fatalf("read error = %v", err)
	}
}

func TestPermissionError(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "Permission denied: /dev/i2c-7", err: errors.New("exit status 1")}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7, monitor.Brightness)
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

func TestDetectWithoutDisplays(t *testing.T) {
	for _, reply := range []response{
		{out: ""},
		{out: "No displays found"},
		{out: "No displays found", err: errors.New("exit status 1")},
	} {
		f := &fakeRunner{responses: []response{reply}}
		displays, err := (&Backend{run: f}).Detect(context.Background())
		if err != nil || len(displays) != 0 {
			t.Fatalf("response=%+v: displays=%v, err=%v", reply, displays, err)
		}
	}
}

func TestDetectDoesNotHideOtherErrors(t *testing.T) {
	for _, output := range []string{"Permission denied: /dev/i2c-7", "No /dev/i2c devices exist. No displays found", "Input/output error\nNo displays found"} {
		f := &fakeRunner{responses: []response{{out: output, err: errors.New("exit status 1")}}}
		if _, err := (&Backend{run: f}).Detect(context.Background()); err == nil {
			t.Fatalf("output %q should be an error", output)
		}
	}
}

func TestMalformedReadIsError(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "something unexpected"}}}
	_, err := (&Backend{run: f}).Read(context.Background(), 7, monitor.Brightness)
	if err == nil {
		t.Fatal("expected malformed response error")
	}
}

func TestReadVolumeCNCAndWrite(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "VCP 62 CNC x00 x64 x00 x32\n"}, {}}}
	b := &Backend{run: f}
	level, err := b.Read(context.Background(), 7, monitor.Volume)
	if err != nil || level != (monitor.Level{Current: 50, Maximum: 100}) {
		t.Fatalf("volume = %+v, err = %v", level, err)
	}
	if err := b.Write(context.Background(), 7, monitor.Volume, 75); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.commands[1], []string{"setvcp", "62", "75", "--bus", "7"}) {
		t.Fatalf("write command = %v", f.commands[1])
	}
}

func TestContrastAndUnsupportedFeature(t *testing.T) {
	f := &fakeRunner{responses: []response{{out: "VCP 12 C 50 100\n"}, {out: "VCP 62 ERR\n"}}}
	b := &Backend{run: f}
	if value, err := b.Read(context.Background(), 7, monitor.Contrast); err != nil || value.Current != 50 {
		t.Fatalf("contrast = %v, err = %v", value, err)
	}
	if _, err := b.Read(context.Background(), 7, monitor.Volume); !errors.Is(err, monitor.ErrUnsupported) {
		t.Fatalf("unsupported volume error = %v", err)
	}
}
