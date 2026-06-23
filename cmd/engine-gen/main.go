package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seargo/seargo/internal/engine/porting"
)

func main() {
	searxngPath := flag.String("searxng", "", "Path to SearXNG source directory")
	outputDir := flag.String("output", "generated", "Output directory for generated Go files")
	baseType := flag.String("base", "", "Filter by base type (xpath, json_engine, mediawiki, custom)")
	limit := flag.Int("limit", 0, "Max engines to generate (0 = all)")
	engineName := flag.String("engine", "", "Generate a single engine by name")
	singleFile := flag.String("file", "", "Generate from a single Python file")
	googleOutput := flag.String("google-output", "data/engine_traits.json", "Output path for `engine-gen google`")

	flag.Parse()

	if len(flag.Args()) > 0 && flag.Args()[0] == "google" {
		if err := runGoogleFetch(*googleOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching google traits: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated google traits in %s\n", *googleOutput)
		return
	}

	if *singleFile != "" {
		generateSingle(*singleFile, *outputDir)
		return
	}

	if *engineName != "" && *searxngPath != "" {
		generateNamed(*searxngPath, *engineName, *outputDir)
		return
	}

	if *searxngPath != "" {
		runSmoke(*searxngPath, *outputDir, *baseType, *limit)
		return
	}

	fmt.Println("Usage: engine-gen --searxng <path> [--output <dir>] [--base <type>] [--limit <n>]")
	fmt.Println("       engine-gen --engine <name> --searxng <path>")
	fmt.Println("       engine-gen --file <python_file>")
	fmt.Println("       engine-gen google [--google-output <path>]")
	flag.PrintDefaults()
}

func generateSingle(filePath, outputDir string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	name := filepath.Base(filePath)
	name = name[:len(name)-len(filepath.Ext(name))]

	result, err := porting.GenerateSkeleton(name, string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating skeleton: %v\n", err)
		os.Exit(1)
	}

	os.MkdirAll(outputDir, 0755)
	outPath := filepath.Join(outputDir, name+".go")
	if err := os.WriteFile(outPath, []byte(result.GoCode), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s -> %s (base: %s)\n", name, outPath, result.BaseType)
}

func generateNamed(searxngPath, engineName, outputDir string) {
	pyPath := filepath.Join(searxngPath, "searx", "engines", engineName+".py")
	generateSingle(pyPath, outputDir)
}

func runSmoke(searxngPath, outputDir, baseType string, limit int) {
	os.MkdirAll(outputDir, 0755)

	results := porting.RunSmoke(porting.SmokeTestConfig{
		SearxngPath: searxngPath,
		OutputDir:   outputDir,
		BaseType:    baseType,
		Limit:       limit,
	})

	success := 0
	failed := 0
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(os.Stderr, "FAIL %s: %s\n", r.Name, r.Error)
			failed++
		} else {
			fmt.Printf("OK   %s -> %s [%s]\n", r.Name, r.Path, r.Output)
			success++
		}
	}

	fmt.Printf("\n%d generated, %d failed\n", success, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
