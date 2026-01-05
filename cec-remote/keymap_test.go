package main

import (
	"testing"
)

func TestGetKeyMap(t *testing.T) {
	// This test assumes /usr/include/linux/input-event-codes.h exists and is readable.
	keyMap := getKeyMap()
	if len(keyMap) == 0 {
		t.Error("Expected non-empty keyMap, got empty map")
	}
	// Check for a common key
	if val, ok := keyMap["KEY_A"]; !ok || val <= 0 {
		t.Error("Expected KEY_A to be present with a positive value")
	}
	// Check for another common key
	if val, ok := keyMap["KEY_B"]; !ok || val <= 0 {
		t.Error("Expected KEY_B to be present with a positive value")
	}
	// Check that a non-existent key is not present
	if _, ok := keyMap["KEY_DOES_NOT_EXIST"]; ok {
		t.Error("Did not expect KEY_DOES_NOT_EXIST to be present in keyMap")
	}
}

func TestGetKeyCode(t *testing.T) {
	keyMap = getKeyMap() // ensure keyMap is initialized
	tests := []struct {
		input    string
		wantCode int
		wantOK   bool
	}{
		{"a", keyMap["KEY_A"], true},
		{"b", keyMap["KEY_B"], true},
		{"space", keyMap["KEY_SPACE"], true},
		{"does_not_exist", 0, false},
	}
	for _, tt := range tests {
		gotCode, gotOK := getKeyCode(tt.input)
		if gotOK != tt.wantOK || (gotOK && gotCode != tt.wantCode) {
			t.Errorf("getKeyCode(%q) = (%d, %v), want (%d, %v)", tt.input, gotCode, gotOK, tt.wantCode, tt.wantOK)
		}
	}
}
