package ui

import "testing"

func TestCursorTimeoutOptionParsing(t *testing.T) {
	for _, tc := range []struct {
		data string
		want float64
		ok   bool
	}{
		{`{"float":0,"set":false}`, 0, true},
		{`{"float":0.5,"set":true}`, 0.5, true},
		{`{"bool":true}`, 0, false},
		{`{"float":-1}`, 0, false},
		{`garbage`, 0, false},
	} {
		got, ok := parseCursorTimeout([]byte(tc.data))
		if ok != tc.ok || !cursorTimeoutEqual(got, tc.want) {
			t.Fatalf("parseCursorTimeout(%q)=(%v,%v), want (%v,%v)", tc.data, got, ok, tc.want, tc.ok)
		}
	}
}

func TestCursorTimeoutIsOnlyForMarchineWindow(t *testing.T) {
	for _, tc := range []struct {
		data string
		want bool
	}{
		{`{"initialClass":"marchine"}`, true},
		{`{"initialClass":"Marchine","class":"foot"}`, true},
		{`{"initialClass":"foot","class":"marchine"}`, false},
		{`{"initialClass":"other"}`, false},
		{`{}`, false},
	} {
		if got := activeWindowIsMarchine([]byte(tc.data)); got != tc.want {
			t.Fatalf("activeWindowIsMarchine(%q)=%v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestCursorTimeoutFocusEvents(t *testing.T) {
	if !cursorFocusEvent("activewindowv2>>0x123") || !cursorFocusEvent("workspace>>2") || !cursorFocusEvent("configreloaded>>") {
		t.Fatal("relevant focus event ignored")
	}
	if cursorFocusEvent("windowtitle>>new title") {
		t.Fatal("unrelated event marked relevant")
	}
}
