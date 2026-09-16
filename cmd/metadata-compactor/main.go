package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	metadatafiltering "github.com/open-resource-discovery/metadata-compactor-golang"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

type CommandLine struct {
	Input string
	Rules string
	Out   string
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

func rules(path string, result *model.Ruleset) *model.Ruleset {
	must("", json.Unmarshal(must(os.ReadFile(path)), result))

	return result
}

func options(args []string, stderr io.Writer) (CommandLine, error) {
	var result CommandLine

	flags := flag.NewFlagSet("metadata-filtering", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.Input, "i", "", "path to the CSN JSON file to filter (required)")
	flags.StringVar(&result.Input, "input", "", "path to the CSN JSON file to filter (required)")
	flags.StringVar(&result.Rules, "r", "", "path to the JSON filtering rules file (required)")
	flags.StringVar(&result.Rules, "rules", "", "path to the JSON filtering rules file (required)")
	flags.StringVar(&result.Out, "o", "", "path for filtered CSN output (default: stdout)")
	flags.StringVar(&result.Out, "output", "", "path for filtered CSN output (default: stdout)")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s -i <input.json> -r <rules.json> [-o <output.json>]\n", flags.Name())
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

	return result, nil
}

func processor(options CommandLine) *metadatafiltering.Processor {
	return metadatafiltering.Create(rules(options.Rules, &model.Ruleset{}))
}

func main() {
	options := must(options(os.Args[1:], os.Stderr))

	write(
		output(options),
		processor(options).Process(metadatafiltering.CSN, string(must(os.ReadFile(options.Input)))),
	)
}
