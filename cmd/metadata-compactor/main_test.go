//go:build unit

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestOptionsBaseline(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"baseline omitted", []string{"-i", "input.json", "-r", "rules.json"}, ""},
		{"short baseline flag", []string{"-i", "input.json", "-r", "rules.json", "-b", "base.json"}, "base.json"},
		{"long baseline flag", []string{"-i", "input.json", "-r", "rules.json", "--baseline", "base.json"}, "base.json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := options(tc.args, &bytes.Buffer{})
			if err != nil {
				t.Fatalf("options() error = %v", err)
			}
			if got.Baseline != tc.want {
				t.Errorf("Baseline = %q, want %q", got.Baseline, tc.want)
			}
		})
	}
}

func TestBaselines(t *testing.T) {
	t.Run("omitted baseline does not read a file", func(t *testing.T) {
		if got := baselines(CommandLine{}); len(got) != 0 {
			t.Errorf("baselines() = %v, want empty", got)
		}
	})

	t.Run("reads configured baseline", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "baseline.json")
		if err := os.WriteFile(path, []byte(`{"definitions":{}}`), 0o600); err != nil {
			t.Fatal(err)
		}

		got := baselines(CommandLine{Baseline: path})
		if len(got) != 1 || got[0] != `{"definitions":{}}` {
			t.Errorf("baselines() = %q, want one baseline document", got)
		}
	})
}
