//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

func TestCLIProcessesDocuments(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "metadata-compactor")
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if output, err := exec.Command(goCommand, "build", "-o", executable, ".").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	writeFile := func(name, content string) string {
		t.Helper()
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	input := writeFile("metadata.edmx", `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="Global"/><Annotation Term="One" String="current"/><Annotation Term="Two"/><Annotation Term="Remove"/></EntityType></Schema>`)
	rules := writeFile("rules.json", `{"edmx":{"preserve":["Global"]}}`)
	first := writeFile("first.edmx", `<Schema><Annotations Target="Demo.Product"><Annotation Term="One" String="old"/></Annotations></Schema>`)
	args := []string{"-f", "edmx", "-i", input, "-r", rules, "-b", first}
	assertOutput := func(t *testing.T, content string) {
		t.Helper()
		for _, term := range []string{"Global", "One"} {
			if !strings.Contains(content, "Term=\""+term+"\"") {
				t.Fatalf("missing %s in %s", term, content)
			}
		}
		if strings.Contains(content, "Remove") || strings.Contains(content, "Two") || strings.Contains(content, "old") || !strings.Contains(content, "current") {
			t.Fatalf("incorrect output: %s", content)
		}
	}
	for _, tc := range []struct {
		format, input, first, expected string
	}{
		{
			format:   "csn",
			input:    `{"definitions":{"Entity":{"kind":"entity","@One":"current","@Two":true,"@Remove":true,"elements":{}}}}`,
			first:    `{"definitions":{"Entity":{"kind":"entity","@One":"old"}}}`,
			expected: `{"definitions":{"Entity":{"kind":"entity","@One":"current","elements":{}}}}`,
		},
		{
			format:   "oas",
			input:    `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0","x-one":"current","x-two":true,"x-remove":true},"paths":{}}`,
			first:    `{"info":{"x-one":"old"}}`,
			expected: `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0","x-one":"current"},"paths":{}}`,
		},
	} {
		t.Run(tc.format+" baseline", func(t *testing.T) {
			input := writeFile(tc.format+".json", tc.input)
			rules := writeFile(tc.format+"-rules.json", `{}`)
			first := writeFile(tc.format+"-first.json", tc.first)
			output, err := exec.Command(executable, "-f", tc.format, "-i", input, "-r", rules, "-b", first).CombinedOutput()
			if err != nil {
				t.Fatalf("CLI: %v\n%s", err, output)
			}
			testutils.AssertDeepEquals(t, oj.MustParseString(tc.expected), oj.MustParseString(string(output)))
		})
	}
	t.Run("edmx stdout", func(t *testing.T) {
		output, err := exec.Command(executable, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("CLI: %v\n%s", err, output)
		}
		assertOutput(t, string(output))
	})
	t.Run("output file", func(t *testing.T) {
		path := filepath.Join(directory, "compacted.edmx")
		output, err := exec.Command(executable, append(args, "-o", path)...).CombinedOutput()
		if err != nil || len(output) != 0 {
			t.Fatalf("CLI: %v\n%s", err, output)
		}
		result, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		assertOutput(t, string(result))
	})
	t.Run("invalid XML does not overwrite output", func(t *testing.T) {
		invalid := writeFile("invalid.edmx", "<unclosed>")
		path := writeFile("existing.edmx", "existing content")
		output, err := exec.Command(executable, "-f", "edmx", "-i", invalid, "-r", rules, "-o", path).CombinedOutput()
		if err == nil || !strings.Contains(string(output), "EDMX input") || strings.Contains(string(output), "panic:") {
			t.Fatalf("expected readable XML error: %v, %s", err, output)
		}
		result, err := os.ReadFile(path)
		if err != nil || string(result) != "existing content" {
			t.Fatalf("output overwritten: %q, %v", result, err)
		}
	})
}
