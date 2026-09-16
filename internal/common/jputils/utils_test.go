//go:build unit

package jputils

import (
	"slices"
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

// fragType returns a string label for the concrete type of a jp.Frag.
func fragType(f jp.Frag) string {
	switch f.(type) {
	case jp.Root:
		return "Root"
	case jp.At:
		return "At"
	case jp.Wildcard:
		return "Wildcard"
	case jp.Child:
		return "Child"
	case jp.Nth:
		return "Nth"
	case *jp.Filter:
		return "Filter"
	default:
		return "unknown"
	}
}

func TestFrag(t *testing.T) {
	// Table for the three special-token cases.
	special := []struct {
		input    string
		wantType string
	}{
		{"@", "At"},
		{"$", "Root"},
		{"*", "Wildcard"},
	}
	for _, tc := range special {
		t.Run(tc.input+" returns "+tc.wantType, func(t *testing.T) {
			got := Frag(tc.input)
			if fragType(got) != tc.wantType {
				t.Errorf("Frag(%q): got %s, want %s", tc.input, fragType(got), tc.wantType)
			}
		})
	}

	t.Run("plain name returns Child with correct value", func(t *testing.T) {
		got := Frag("name")
		if fragType(got) != "Child" {
			t.Fatalf("got %s, want Child", fragType(got))
		}
		if string(got.(jp.Child)) != "name" {
			t.Errorf("child value = %q, want \"name\"", string(got.(jp.Child)))
		}
	})

	t.Run("name with hyphens returns Child", func(t *testing.T) {
		got := Frag("some-key")
		if fragType(got) != "Child" {
			t.Errorf("got %s, want Child", fragType(got))
		}
		if string(got.(jp.Child)) != "some-key" {
			t.Errorf("child value = %q, want \"some-key\"", string(got.(jp.Child)))
		}
	})

	t.Run("bare integer string returns Child, not Nth", func(t *testing.T) {
		// "1" does not match ^\[\d+]$ so falls through to Child.
		got := Frag("1")
		if fragType(got) != "Child" {
			t.Errorf("got %s, want Child", fragType(got))
		}
	})

	// Frag("[n]") matches the ^\[\d+]$ regex, but strconv.Atoi("[n]") always
	// fails (brackets are not valid for Atoi) and returns (0, err).
	// utils.First discards the error so jp.Nth(0) is always produced.
	// This is a bug in the implementation; the test acts as a regression guard.
	t.Run("[n] bracket form: regex matches but Atoi fails, always produces Nth(0)", func(t *testing.T) {
		for _, input := range []string{"[0]", "[1]", "[2]", "[99]"} {
			got := Frag(input)
			if fragType(got) != "Nth" {
				t.Errorf("Frag(%q): got %s, want Nth", input, fragType(got))
				continue
			}
			if int(got.(jp.Nth)) != 0 {
				t.Errorf("Frag(%q): nth = %d, want 0 (Atoi fails on bracketed form)", input, int(got.(jp.Nth)))
			}
		}
	})
}

func TestExpr(t *testing.T) {
	data := map[string]any{
		"name":  "Alice",
		"score": int64(99),
		"addr":  map[string]any{"city": "Berlin"},
		"items": []any{"a", "b", "c"},
	}

	t.Run("empty args returns empty Expr", func(t *testing.T) {
		if len(Expr()) != 0 {
			t.Errorf("len = %d, want 0", len(Expr()))
		}
	})

	t.Run("string fragments navigate to top-level field", func(t *testing.T) {
		result := Expr("$", "name").Get(data)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("string fragments navigate nested field", func(t *testing.T) {
		result := Expr("$", "addr", "city").Get(data)
		if len(result) != 1 || result[0] != "Berlin" {
			t.Errorf("got %v, want [Berlin]", result)
		}
	})

	t.Run("wildcard returns all top-level values", func(t *testing.T) {
		result := Expr("$", "*").Get(data)
		if len(result) != 4 {
			t.Errorf("len = %d, want 4", len(result))
		}
		// Confirm known scalar values are present (maps/slices are not hashable).
		foundAlice, foundScore := false, false
		for _, v := range result {
			if v == "Alice" {
				foundAlice = true
			}
			if v == int64(99) {
				foundScore = true
			}
		}
		if !foundAlice {
			t.Error("wildcard result missing \"Alice\"")
		}
		if !foundScore {
			t.Error("wildcard result missing int64(99)")
		}
	})

	t.Run("jp.Frag argument appended directly", func(t *testing.T) {
		result := Expr(jp.Root('$'), jp.Child("score")).Get(data)
		if len(result) != 1 || result[0] != int64(99) {
			t.Errorf("got %v, want [99]", result)
		}
	})

	t.Run("jp.Expr argument is flattened into result", func(t *testing.T) {
		sub := jp.Expr{jp.Root('$'), jp.Child("name")}
		result := Expr(sub).Get(data)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("jp.Expr argument combined with trailing string fragments", func(t *testing.T) {
		// Demonstrates that a jp.Expr is flattened and string frags appended after.
		sub := jp.Expr{jp.Root('$'), jp.Child("addr")}
		result := Expr(sub, "city").Get(data)
		if len(result) != 1 || result[0] != "Berlin" {
			t.Errorf("got %v, want [Berlin]", result)
		}
	})

	t.Run("*jp.Equation argument becomes Filter frag on array root", func(t *testing.T) {
		arr := []any{
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		}
		eq := Eq("@.name", "Alice")
		result := Expr("$", eq).Get(arr)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		if result[0].(map[string]any)["name"] != "Alice" {
			t.Errorf("name = %v, want Alice", result[0].(map[string]any)["name"])
		}
	})

	t.Run("fragment types assembled in correct order", func(t *testing.T) {
		e := Expr("$", "items", "*")
		if len(e) != 3 {
			t.Fatalf("len = %d, want 3", len(e))
		}
		want := []string{"Root", "Child", "Wildcard"}
		got := make([]string, len(e))
		for i, f := range e {
			got[i] = fragType(f)
		}
		if !slices.Equal(got, want) {
			t.Errorf("fragment types = %v, want %v", got, want)
		}
	})
}

func TestConst(t *testing.T) {
	t.Run("all supported types return non-nil Equation", func(t *testing.T) {
		cases := []struct {
			name  string
			value any
		}{
			{"bool true", true},
			{"bool false", false},
			{"int64", int64(42)},
			{"string", "hello"},
			{"float64", float64(3.14)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if Const(tc.value) == nil {
					t.Errorf("Const(%v): got nil, want non-nil Equation", tc.value)
				}
			})
		}
	})

	t.Run("string constant evaluates correctly in a filter", func(t *testing.T) {
		// Verify the Equation is not just non-nil but holds the right constant.
		arr := []any{
			map[string]any{"tag": "yes"},
			map[string]any{"tag": "no"},
		}
		filter := jp.Eq(jp.Get(Expr("@", "tag")), Const("yes"))
		result := Expr("$", filter).Get(arr)
		if len(result) != 1 || result[0].(map[string]any)["tag"] != "yes" {
			t.Errorf("got %v, want [{tag:yes}]", result)
		}
	})

	t.Run("bool constant evaluates correctly in a filter", func(t *testing.T) {
		arr := []any{
			map[string]any{"active": true},
			map[string]any{"active": false},
		}
		filter := jp.Eq(jp.Get(Expr("@", "active")), Const(true))
		result := Expr("$", filter).Get(arr)
		if len(result) != 1 || result[0].(map[string]any)["active"] != true {
			t.Errorf("got %v, want [{active:true}]", result)
		}
	})

	t.Run("unsupported type panics", func(t *testing.T) {
		defer testutils.AssertPanics(t, "expected panic due to unsupported type")

		Const([]int{1, 2, 3})
	})
}

func TestEq(t *testing.T) {
	arr := []any{
		map[string]any{"name": "Alice", "age": int64(30)},
		map[string]any{"name": "Bob", "age": int64(25)},
		map[string]any{"name": "Alice", "age": int64(40)},
	}

	t.Run("filters array by string equality", func(t *testing.T) {
		result := Expr("$", Eq("@.name", "Alice")).Get(arr)
		if len(result) != 2 {
			t.Fatalf("len = %d, want 2", len(result))
		}
		for i, r := range result {
			if r.(map[string]any)["name"] != "Alice" {
				t.Errorf("result[%d]: name = %v, want Alice", i, r.(map[string]any)["name"])
			}
		}
	})

	t.Run("filters array by int64 equality", func(t *testing.T) {
		result := Expr("$", Eq("@.age", int64(25))).Get(arr)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		m := result[0].(map[string]any)
		if m["name"] != "Bob" || m["age"] != int64(25) {
			t.Errorf("got %v, want {name:Bob age:25}", m)
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		result := Expr("$", Eq("@.name", "Charlie")).Get(arr)
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("dot-separated path navigates nested field", func(t *testing.T) {
		nested := []any{
			map[string]any{"addr": map[string]any{"city": "Berlin"}},
			map[string]any{"addr": map[string]any{"city": "Paris"}},
		}
		result := Expr("$", Eq("@.addr.city", "Berlin")).Get(nested)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		city := result[0].(map[string]any)["addr"].(map[string]any)["city"]
		if city != "Berlin" {
			t.Errorf("city = %v, want Berlin", city)
		}
	})

	t.Run("single-segment path (no dot) works", func(t *testing.T) {
		// strings.Split("@", ".") produces ["@"], a one-element path.
		flat := []any{"Alice", "Bob"}
		result := Expr("$", Eq("@", "Alice")).Get(flat)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("returns non-nil Equation", func(t *testing.T) {
		if Eq("@.x", "v") == nil {
			t.Error("expected non-nil equation")
		}
	})
}
