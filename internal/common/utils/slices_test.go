//go:build unit

package utils

import (
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("doubles each integer", func(t *testing.T) {
		got := Map([]int{1, 2, 3}, func(_ int, v int) int { return v * 2 })
		want := []int{2, 4, 6}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("index is passed correctly", func(t *testing.T) {
		got := Map([]string{"a", "b", "c"}, func(i int, _ string) int { return i })
		want := []int{0, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("type change: bool to string", func(t *testing.T) {
		got := Map([]bool{true, false, true}, func(_ int, v bool) string {
			if v {
				return "yes"
			}
			return "no"
		})
		want := []string{"yes", "no", "yes"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		got := Map([]int{}, func(_ int, v int) int { return v })
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := Map([]int{7}, func(_ int, v int) int { return v + 1 })
		if !slices.Equal(got, []int{8}) {
			t.Errorf("got %v, want [8]", got)
		}
	})

	t.Run("mapper receives correct index and value for each element", func(t *testing.T) {
		// verifies both idx and v are forwarded correctly by combining them
		got := Map([]int{10, 20, 30}, func(i int, v int) int { return i*100 + v })
		want := []int{10, 120, 230}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected []int
		want     bool
	}{
		{"value present in list", 2, []int{1, 2, 3}, true},
		{"value not in list", 5, []int{1, 2, 3}, false},
		{"empty expected list returns false", 1, []int{}, false},
		{"single matching expected", 7, []int{7}, true},
		{"single non-matching expected", 7, []int{8}, false},
		{"first element matches", 1, []int{1, 2, 3}, true},
		{"last element matches", 3, []int{1, 2, 3}, true},
		{"zero value not in list", 0, []int{1, 2, 3}, false},
		{"zero value in list", 0, []int{1, 0, 3}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OneOf(tc.value, tc.expected...)
			if got != tc.want {
				t.Errorf("OneOf(%v, %v) = %v, want %v", tc.value, tc.expected, got, tc.want)
			}
		})
	}

	t.Run("string values", func(t *testing.T) {
		if !OneOf("banana", "apple", "banana", "cherry") {
			t.Error("expected true")
		}
	})
}

func TestNone(t *testing.T) {
	tests := []struct {
		name      string
		elements  []int
		predicate func(int) bool
		want      bool
	}{
		{"no elements match", []int{1, 3, 5}, func(v int) bool { return v%2 == 0 }, true},
		{"an element matches", []int{1, 2, 3}, func(v int) bool { return v%2 == 0 }, false},
		{"empty input returns true", []int{}, func(int) bool { return true }, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := None(tc.elements, tc.predicate); got != tc.want {
				t.Errorf("None(%v) = %v, want %v", tc.elements, got, tc.want)
			}
		})
	}

	t.Run("stops after the first match", func(t *testing.T) {
		calls := 0
		got := None([]int{1, 2, 3}, func(v int) bool {
			calls++
			return v == 2
		})
		if got {
			t.Error("expected false")
		}
		if calls != 2 {
			t.Errorf("predicate called %d times, want 2", calls)
		}
	})
}

func TestLast(t *testing.T) {
	tests := []struct {
		name     string
		elements []int
		want     int
	}{
		{"returns final element", []int{1, 2, 3}, 3},
		{"returns the only element", []int{7}, 7},
		{"returns a zero value when it is final", []int{1, 0}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Last(tc.elements); got != tc.want {
				t.Errorf("Last(%v) = %v, want %v", tc.elements, got, tc.want)
			}
		})
	}

	t.Run("supports string elements", func(t *testing.T) {
		if got := Last([]string{"first", "last"}); got != "last" {
			t.Errorf("Last returned %q, want %q", got, "last")
		}
	})
}
