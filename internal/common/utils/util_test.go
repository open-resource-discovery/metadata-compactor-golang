//go:build unit

package utils

import (
	"testing"
)

func TestFirst(t *testing.T) {
	t.Run("returns first of two args", func(t *testing.T) {
		if got := First("hello", "world"); got != "hello" {
			t.Errorf("got %q, want \"hello\"", got)
		}
	})

	t.Run("returns first int", func(t *testing.T) {
		if got := First(42, "ignored"); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("single argument", func(t *testing.T) {
		if got := First(true); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("works with multi-return function result", func(t *testing.T) {
		multiReturn := func() (int, error) { return 7, nil }
		if got := First(multiReturn()); got != 7 {
			t.Errorf("got %d, want 7", got)
		}
	})
}

func TestTernary(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		first     string
		second    string
		want      string
	}{
		{"true returns first", true, "yes", "no", "yes"},
		{"false returns second", false, "yes", "no", "no"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Ternary(tc.condition, tc.first, tc.second)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("integer values", func(t *testing.T) {
		if got := Ternary(1 > 2, 100, 200); got != 200 {
			t.Errorf("got %d, want 200", got)
		}
	})

	t.Run("bool values", func(t *testing.T) {
		if got := Ternary(true, true, false); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("pointer values", func(t *testing.T) {
		a, b := 1, 2
		got := Ternary(false, &a, &b)
		if got != &b {
			t.Error("expected pointer to b")
		}
	})
}
