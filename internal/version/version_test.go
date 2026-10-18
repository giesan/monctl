package version

import "testing"

func TestWindowTitle(t *testing.T) {
	original := Value
	defer func() { Value = original }()
	for _, tc := range []struct {
		value string
		want  string
	}{
		{value: "dev", want: "Monitor Control dev"},
		{value: "v1.2.3", want: "Monitor Control v1.2.3"},
	} {
		Value = tc.value
		if got := WindowTitle(); got != tc.want {
			t.Fatalf("window title for %q = %q, want %q", tc.value, got, tc.want)
		}
	}
}
