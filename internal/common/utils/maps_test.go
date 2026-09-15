//go:build unit

package utils

import (
	"testing"
)

func TestContainsKey(t *testing.T) {
	m := map[string]int{"x": 1, "y": 2}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"present key", "x", true},
		{"another present key", "y", true},
		{"absent key", "z", false},
		{"empty string key absent", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsKey(m, tc.key)
			if got != tc.want {
				t.Errorf("ContainsKey(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}

	t.Run("empty map always false", func(t *testing.T) {
		if ContainsKey(map[string]int{}, "anything") {
			t.Error("expected false for empty map")
		}
	})

	t.Run("zero-value key present", func(t *testing.T) {
		m2 := map[string]int{"": 99}
		if !ContainsKey(m2, "") {
			t.Error("expected true for zero-value key")
		}
	})
}
