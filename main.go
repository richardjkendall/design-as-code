package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Resource struct {
	resourceType       string
	resourceName       string
	resourceAttributes map[string]string
}

func DecodeBody(body *hcl.BodyContent, resourceType string) ([]Resource, hcl.Diagnostics) {

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
	patternsLibraryFile := flag.String("patternlib", "patterns.hcl", "Path to the file containing the list of patterns to use for matching.")
	solutionDescriptor := flag.String("app", "app.hcl", "Path to the file containing the list of patterns to use for matching.")
	solveMode := flag.String("solvefor", "priority", "What solution mode should we use.")
	debugLog := flag.Bool("debug", false, "Should we log verbose messages for debugging?")
	flag.Parse()

	if *debugLog {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Running...")
	log.WithFields(log.Fields{
		"patternLibraryFile": *patternsLibraryFile,
	}).Info("Patterns library file")
	log.WithFields(log.Fields{
		"solutionDescriptorFile": *solutionDescriptor,
	}).Info("Solution descriptor file")

	log.Info("Loading patterns...")
	var patterns Patterns
	patterns, err := LoadPatternLibrary(*patternsLibraryFile)
	if err != nil {
		log.Fatal("Failed to load patterns: ", err)
	}
	log.Info("Got ", len(patterns.PatternSet), " patterns")

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

		log.Info("Got ", len(resources), " resources")
		for _, resource := range resources {
			log.WithFields(log.Fields{
				"resourceType": resource.resourceType,
				"resourceName": resource.resourceName,
			}).Debug("Resource object")
			for key, value := range resource.resourceAttributes {
				log.WithFields(log.Fields{
					"resource": resource.resourceType + "/" + resource.resourceName,
					"variable": key,
					"value":    value,
				}).Debug("Variable on resource")
			}
		}

		log.Info("Doing intial pattern match")
		matched, unmatched := MatchPatternsToSolution(resources, patterns.PatternSet)
		log.WithFields(log.Fields{
			"matched":   len(matched),
			"unmatched": len(unmatched),
		}).Info("Matched patterns")

		log.WithFields(log.Fields{
			"solveMode": *solveMode,
		}).Info("Running solver")
		var solution []MatchedPattern
		var unmatchedAfterSolution []string

		if *solveMode == "priority" {
			solution, unmatchedAfterSolution = SolveForPriority(matched, resources)
		}
		if *solveMode == "max" {
			solution, unmatchedAfterSolution = SolvForMaxCoverage(matched, resources)
		}
		log.WithFields(log.Fields{
			"matched":   len(solution),
			"unmatched": len(unmatchedAfterSolution),
		}).Info("Solver has run")

		fmt.Print("\nMatched patterns\n\n")
		PrintTextPatternTable(solution)

		if len(unmatchedAfterSolution) == 0 {
			fmt.Print("\nNo unmatched resources.\n")
		} else {
			fmt.Print("\nUmatched resources:\n\n")
			PrintTextResourceTable(unmatchedAfterSolution)
		}

	}

}
