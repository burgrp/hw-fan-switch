package fanswitch

import "testing"

func TestNormalizedDuty(t *testing.T) {
	tests := []struct {
		name  string
		value int32
		null  bool
		want  int32
	}{
		{name: "minimum", value: -1, want: 0},
		{name: "unchanged", value: 30, want: 30},
		{name: "maximum", value: 101, want: 100},
		{name: "null stops fan", value: 75, null: true, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizedDuty(test.value, test.null); got != test.want {
				t.Fatalf("normalizedDuty(%d, %v) = %d, want %d", test.value, test.null, got, test.want)
			}
		})
	}
}
