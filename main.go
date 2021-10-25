package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Resource struct {
	resourceType       string
	resourceName       string
	resourceAttributes map[string]string
}

func DecodeBody(body *hcl.BodyContent, resourceType string) ([]Resource, hcl.Diagnostics) {
	fmt.Println("Starting DecodeBody...")

	var resources []Resource

	ctx := &hcl.EvalContext{}
	for _, block := range body.Blocks.OfType("resource") {
		resourceType := block.Labels[0]

		var resource Resource
		attributes := make(map[string]string)

		resource.resourceName = block.Labels[1]
		resource.resourceType = resourceType

		schema := schemaMap[resourceType]
		contents, diagnostics := block.Body.Content(&schema)
		if diagnostics != nil && diagnostics.HasErrors() {
			return nil, diagnostics
		}

		for _, attribute := range contents.Attributes {
			val, _ := attribute.Expr.Value(ctx)
			attributes[attribute.Name] = convertValueToString(val)
		}

		resource.resourceAttributes = attributes

		resources = append(resources, resource)
	}

	return resources, nil
}

func main() {

	// need to get the command line parameters
	patternsLibraryFile := flag.String("patternlib", "patterns.hcl", "Path to the file containing the list of patterns to use for matching.  Defaults to 'patterns.hcl'")
	solutionDescriptor := flag.String("app", "app.hcl", "Path to the file containing the list of patterns to use for matching.  Defaults to 'app.hcl'")
	solveMode := flag.String("mode", "priority", "What solution mode should we use.")
	flag.Parse()

	fmt.Println("Running...")
	fmt.Printf("Patterns library file:    %s\n", *patternsLibraryFile)
	fmt.Printf("Solution descriptor file: %s\n", *solutionDescriptor)

	fmt.Println("Loading patterns...")
	var patterns Patterns
	patterns, err := LoadPatternLibrary(*patternsLibraryFile)
	if err != nil {
		log.Fatalf("Failed to load patterns: %s", err)
	}
	fmt.Printf("Got %d patterns\n\n", len(patterns.PatternSet))

	p := hclparse.NewParser()

	wr := hcl.NewDiagnosticTextWriter(
		os.Stdout, // writer to send messages to
		p.Files(), // the parser's file cache, for source snippets
		78,        // wrapping width
		true,      // generate colored/highlighted output
	)

	_, diagnostics := p.ParseHCLFile(*solutionDescriptor)
	if diagnostics != nil && diagnostics.HasErrors() {
		wr.WriteDiagnostics(diagnostics)
	}

	for _, file := range p.Files() {
		contents, diagnostics := file.Body.Content(solutionSchema)
		if diagnostics != nil && diagnostics.HasErrors() {
			wr.WriteDiagnostics(diagnostics)
		}

		// call descent parser from here
		resources, diagnostics := DecodeBody(contents, "")
		if diagnostics != nil && diagnostics.HasErrors() {
			wr.WriteDiagnostics(diagnostics)
		}

		fmt.Println("\nGot a set of resources...")
		for _, resource := range resources {
			fmt.Printf("\tResource %s/%s\n", resource.resourceType, resource.resourceName)
			for key, value := range resource.resourceAttributes {
				fmt.Printf("\t\t-> %s = %s\n", key, value)
			}
		}

		fmt.Println("\nAttempting matches...")
		matched, unmatched := MatchPatternsToSolution(resources, patterns.PatternSet)
		fmt.Printf("\nMatched %d patterns & left %d umatched resources\n", len(matched), len(unmatched))

		fmt.Println("\nRunning solver..., mode: ", *solveMode)
		var solution []MatchedPattern
		var unmatchedAfterSolution []string

		if *solveMode == "priority" {
			solution, unmatchedAfterSolution = SolveForPriority(matched, resources)
		}
		if *solveMode == "max" {
			solution, unmatchedAfterSolution = SolvForMaxCoverage(matched, resources)
		}
		fmt.Printf("\nMatched %d patterns & left %d umatched resources\n", len(solution), len(unmatchedAfterSolution))

		for _, mp := range solution {
			fmt.Printf("\tPattern %s / %s\n", mp.Pattern.PatternName, mp.Pattern.Description)
			for _, mr := range mp.Resources {
				fmt.Printf("\t\t-> %s/%s\n", mr.resourceType, mr.resourceName)
			}
		}

		fmt.Printf("\nUmatched resources:\n")
		for _, un := range unmatchedAfterSolution {
			fmt.Printf("\t%s\n", un)
		}

	}

}
