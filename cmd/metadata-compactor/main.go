package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	metadatacompactor "github.com/open-resource-discovery/metadata-compactor-golang"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

type CommandLine struct {
	Input    string
	Format   metadatacompactor.Format
	Rules    string
	Out      string
	Baseline string
}

func load(path string) string {
	return string(must(os.ReadFile(path)))
}

func must[E any](value E, err error) E {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return value
}

func write(writer io.Writer, data string) {
	must(writer.Write([]byte(data)))
}

func output(options CommandLine) io.Writer {
	if options.Out == "" {
		return os.Stdout
	}
	return must(os.Create(options.Out))
}

func baselines(options CommandLine) []string {
	if options.Baseline == "" {
		return nil
	}
	return []string{load(options.Baseline)}
}

func rules(path string, result *model.Ruleset) *model.Ruleset {
	must("", json.Unmarshal(must(os.ReadFile(path)), result))
	return result
}

func options(args []string, stderr io.Writer) (CommandLine, error) {
	var result CommandLine
	var inputFormat string

	flags := flag.NewFlagSet("metadata-compactor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.Input, "i", "", "path to the document to compact (required)")
	flags.StringVar(&result.Input, "input", "", "path to the document to compact (required)")
	flags.StringVar(&inputFormat, "f", "csn", "input format: csn, oas, or edmx")
	flags.StringVar(&inputFormat, "format", "csn", "input format: csn, oas, or edmx")
	flags.StringVar(&result.Rules, "r", "", "path to the JSON compaction rules file (required)")
	flags.StringVar(&result.Rules, "rules", "", "path to the JSON compaction rules file (required)")
	flags.StringVar(&result.Out, "o", "", "path for compacted output (default: stdout)")
	flags.StringVar(&result.Out, "output", "", "path for compacted output (default: stdout)")
	flags.StringVar(&result.Baseline, "b", "", "path to a baseline document")
	flags.StringVar(&result.Baseline, "baseline", "", "path to a baseline document")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s -i <input> -r <rules.json> [-f <csn|oas|edmx>] [-o <output>] [-b <baseline>]\n", flags.Name())
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return result, err
	}
	if flags.NArg() > 0 {
		return result, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if result.Input == "" {
		return result, errors.New("missing required -i argument")
	}
	if result.Rules == "" {
		return result, errors.New("missing required -r argument")
	}
	switch strings.ToLower(inputFormat) {
	case "csn":
		result.Format = metadatacompactor.CSN
	case "oas":
		result.Format = metadatacompactor.OAS
	case "edmx":
		result.Format = metadatacompactor.EDMX
	default:
		return result, fmt.Errorf("unsupported input format %q: expected csn, oas, or edmx", inputFormat)
	}
	return result, nil
}

func main() {
	options := must(options(os.Args[1:], os.Stderr))
	processor := metadatacompactor.Create(rules(options.Rules, &model.Ruleset{}))
	result := must(processor.ProcessE(options.Format, load(options.Input), baselines(options)...))
	write(output(options), result)
}
