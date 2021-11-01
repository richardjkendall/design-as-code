package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func main() {

	// need to get the command line parameters
	patternsLibraryFile := flag.String("patternlib", "patterns.hcl", "Path to the file containing the list of patterns to use for matching.")
	solutionDescriptor := flag.String("app", "app.hcl", "Path to the file containing the list of patterns to use for matching.")
	toolMode := flag.String("mode", "match", "What should the tool do 'match' or 'describe'")
	solveMode := flag.String("solvefor", "priority", "What solution mode should we use.")
	jsonFileOut := flag.String("json", "", "Should we output to json, if so, what file name.")
	debugLog := flag.Bool("debug", false, "Should we log verbose messages for debugging?")
	traceLog := flag.Bool("trace", false, "Should we log verbose messages for debugging?")
	flag.Parse()

	if *debugLog {
		log.SetLevel(log.DebugLevel)
	}

	if *traceLog {
		log.SetLevel(log.TraceLevel)
	}

	if *toolMode != "match" && *toolMode != "describe" {
		log.Fatal("Tool mode (mode) is incorrect, expecting 'match' or 'describe'")
	}

	log.Info("Running...")
	log.WithFields(log.Fields{
		"patternLibraryFile": *patternsLibraryFile,
	}).Info("Patterns library file")
	log.WithFields(log.Fields{
		"solutionDescriptorFile": *solutionDescriptor,
	}).Info("Solution descriptor file")

	log.Info("Loading solution schema...")
	schemas := make(map[string]hcl.BodySchema)
	typemap := make(map[string]map[string]string)
	schemareaderr := ReadSchema(schemas, typemap)
	if schemareaderr != nil {
		log.WithError(schemareaderr).Fatal("Cannot continue")
	}

	log.Info("Loading patterns...")
	var patterns Patterns
	patterns, err := LoadPatternLibrary(*patternsLibraryFile)
	if err != nil {
		log.Fatal("Failed to load patterns: ", err)
	}
	log.WithFields(log.Fields{
		"count": len(patterns.PatternSet),
	}).Info("Loaded pattern library")

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
		resources, app, diagnostics := DecodeBody(contents, "", schemas, typemap)
		if diagnostics != nil && diagnostics.HasErrors() {
			wr.WriteDiagnostics(diagnostics)
			log.Fatal("Unrecoverable error")
			os.Exit(1)
		}

		log.WithFields(log.Fields{
			"count":          len(resources),
			"solutionName":   app.solutionName,
			"solutionNumber": app.solutionNumber,
		}).Info("Solution loaded")

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

		log.WithFields(log.Fields{
			"mode": *toolMode,
		}).Info("Tool mode")

		if *jsonFileOut != "" {
			log.WithFields(log.Fields{
				"jsonFile": *jsonFileOut,
			}).Info("Output mode is JSON")
		}

		if *toolMode == "describe" {
			log.Info("Mode is describe")
			if *jsonFileOut == "" {
				log.Warn("Mode is 'describe', but no JSON file was specified for output")
			} else {
				log.Debug("Getting formatted data for JSON conversion")
				data := ResourcesToStringMap(resources, app)
				log.Debug("Converting data to JSON")
				rows, err := ListToJson(data)
				if err != nil {
					log.WithError(err).Fatal("Error encoding JSON")
				}
				jsonForFile := strings.Join(rows, "\n")
				fmt.Printf("\nJSON: \n%s\n\n", jsonForFile)
				log.Debug("Writing to file")
				writeErr := ioutil.WriteFile(*jsonFileOut, []byte(jsonForFile), 0644)
				if writeErr != nil {
					log.WithError(writeErr).Fatal("Error writing to file")
				}
				log.Debug("File written")
			}
		}

		if *toolMode == "match" {
			log.Info("Doing intial pattern match")
			matched, unmatched := MatchPatternsToSolution(resources, patterns.PatternSet, typemap)
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

			// need to write to JSON if the mode is enabled
			if *jsonFileOut != "" {
				log.Debug("Getting formatted data for JSON conversion")
				data := MatchedPatternsToStringMap(solution, resources, unmatchedAfterSolution, app)
				log.Debug("Converting data to JSON")
				rows, err := ListToJson(data)
				if err != nil {
					log.WithError(err).Fatal("Error encoding JSON")
				}
				jsonForFile := strings.Join(rows, "\n")
				fmt.Printf("\nJSON: \n%s\n\n", jsonForFile)
				log.Debug("Writing to file")
				writeErr := ioutil.WriteFile(*jsonFileOut, []byte(jsonForFile), 0644)
				if writeErr != nil {
					log.WithError(writeErr).Fatal("Error writing to file")
				}
				log.Debug("File written")
			}

		}

	}

}
