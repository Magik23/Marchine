package ui

import "testing"

func TestParseMarchineHyprlandFullscreen(t *testing.T) {
	tests := []struct {
		name string
		json string
		want bool
	}{
		{
			name: "Marchine fullscreen",
			json: `{"initialClass":"marchine","fullscreen":2}`,
			want: true,
		},
		{
			name: "Marchine fullscreened maximized state",
			json: `{"initialClass":"marchine","fullscreen":3}`,
			want: true,
		},
		{
			name: "Marchine windowed",
			json: `{"initialClass":"marchine","fullscreen":0}`,
			want: false,
		},
		{
			name: "Different active window",
			json: `{"initialClass":"mame","fullscreen":2}`,
			want: false,
		},
		{
			name: "Invalid JSON",
			json: `{`,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseMarchineHyprlandFullscreen([]byte(test.json)); got != test.want {
				t.Fatalf("fullscreen=%v want=%v", got, test.want)
			}
		})
	}
}
