//go:build unit || integration

package testutils

import (
	"testing"
)

// Get navigates a nested map[string]any by successive string keys, fataling on type mismatch.
func Get(t *testing.T, doc map[string]any, keys ...string) any {
	t.Helper()

	var cur any = doc
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("Get: expected map at key %q, got %T", k, cur)
		}
		cur = m[k]
	}
	return cur
}
