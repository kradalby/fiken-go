// Command fiken-spec-lint walks api/fiken-openapi.yaml and fails if
// any schema property whose name matches *Date or *At|*Time|*DateTime
// lacks the corresponding OAS `format:` declaration. Catches naive
// datetime strings that would otherwise silently parse as UTC in Go.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultSpec = "api/fiken-openapi.yaml"

type property struct {
	Type   string `yaml:"type"`
	Format string `yaml:"format"`
}

func main() {
	specPath := flag.String("spec", defaultSpec, "path to OAS YAML")
	ignoreFlag := flag.String("ignore", "", "comma-separated field names to skip")
	flag.Parse()

	ignored := map[string]bool{}
	for _, s := range strings.Split(*ignoreFlag, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			ignored[s] = true
		}
	}

	violations, err := run(*specPath, ignored)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fiken-spec-lint: %v\n", err)
		os.Exit(2)
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}
		fmt.Fprintf(os.Stderr, "\n%d unit-format violation(s)\n", len(violations))
		os.Exit(1)
	}
	fmt.Println("ok")
}

func run(specPath string, ignored map[string]bool) ([]string, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}
	var root struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]property `yaml:"properties"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	var out []string
	for schemaName, schema := range root.Components.Schemas {
		for propName, prop := range schema.Properties {
			out = append(out, checkProperty(schemaName, propName, prop, ignored)...)
		}
	}
	return out, nil
}

func checkProperty(schema, name string, p property, ignored map[string]bool) []string {
	if ignored[name] {
		return nil
	}
	low := strings.ToLower(name)
	switch {
	case strings.HasSuffix(low, "datetime"):
		if p.Format != "date-time" {
			return []string{fmt.Sprintf("%s.%s: name ends in DateTime but format=%q (want date-time)", schema, name, p.Format)}
		}
	case strings.HasSuffix(low, "at"), strings.HasSuffix(low, "time"):
		if strings.HasSuffix(low, "date") {
			break
		}
		if p.Type != "string" {
			return nil
		}
		if p.Format != "date-time" {
			return []string{fmt.Sprintf("%s.%s: name suggests datetime but format=%q (want date-time)", schema, name, p.Format)}
		}
	case strings.HasSuffix(low, "date"):
		if p.Type != "string" {
			return nil
		}
		if p.Format != "date" && p.Format != "date-time" {
			return []string{fmt.Sprintf("%s.%s: name ends in Date but format=%q (want date or date-time)", schema, name, p.Format)}
		}
	}
	return nil
}
