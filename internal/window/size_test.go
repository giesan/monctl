package window

import "testing"

func TestFitHeight(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		minimum, natural, workarea int
		want                       int
	}{
		{"fits cards exactly", 90, 282, 900, 282},
		{"scrolls on a small screen", 90, 850, 600, 520},
		{"status increases natural height", 90, 338, 900, 338},
		{"hidden status shrinks again", 90, 282, 900, 282},
		{"unknown workarea", 90, 282, 0, 282},
		{"respects GTK minimum", 190, 400, 220, 190},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := fitHeight(tc.minimum, tc.natural, tc.workarea); got != tc.want {
				t.Fatalf("fitHeight(%d, %d, %d) = %d, want %d", tc.minimum, tc.natural, tc.workarea, got, tc.want)
			}
		})
	}
}
