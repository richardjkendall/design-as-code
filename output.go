package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

func MatchedPatternsToStringMap(matches []MatchedPattern, resources []Resource, unmatched []string, solution Solution) (out []map[string]interface{}) {
	version := time.Now().Unix()
	// deal with the matches first
	for _, match := range matches {
		for _, resource := range match.Resources {
			resourceMap := make(map[string]interface{})
			resourceMap["solutionName"] = solution.solutionName
			resourceMap["solutionNumber"] = solution.solutionNumber
			resourceMap["version"] = version
			resourceMap["resourceName"] = resource.resourceName
			resourceMap["resourceType"] = resource.resourceType
			resourceMap["patternName"] = match.Pattern.PatternName
			resourceMap["patternTarget"] = match.Pattern.Target
			resourceMap["matchesPattern"] = true
			attributes := make([]map[string]interface{}, 0)
			for attributeName, attributeValue := range resource.resourceAttributes {
				if attributeName != "depends_on" {
					attribute := make(map[string]interface{})
					attribute["name"] = attributeName
					attribute["value"] = attributeValue
					attributes = append(attributes, attribute)
				} else {
					resourceMap["dependsOn"] = attributeValue
				}
			}
			resourceMap["attributes"] = attributes
			out = append(out, resourceMap)
		}
	}

	// create map for the unmatched resources
	unmatchedMap := make(map[string]bool)
	for _, resource := range unmatched {
		unmatchedMap[resource] = true
	}
	for _, resource := range resources {
		// check if resource is unmatched
		_, present := unmatchedMap[resource.resourceType+"/"+resource.resourceName]
		if present {
			resourceMap := make(map[string]interface{})
			resourceMap["solutionName"] = solution.solutionName
			resourceMap["solutionNumber"] = solution.solutionNumber
			resourceMap["version"] = version
			resourceMap["resourceName"] = resource.resourceName
			resourceMap["resourceType"] = resource.resourceType
			resourceMap["matchesPattern"] = false
			attributes := make([]map[string]interface{}, 0)
			for attributeName, attributeValue := range resource.resourceAttributes {
				if attributeName != "depends_on" {
					attribute := make(map[string]interface{})
					attribute["name"] = attributeName
					attribute["value"] = attributeValue
					attributes = append(attributes, attribute)
				} else {
					resourceMap["dependsOn"] = attributeValue
				}
			}
			resourceMap["attributes"] = attributes
			out = append(out, resourceMap)
		}
	}

	return
}

func ListToJson(in []map[string]interface{}) ([]string, error) {
	var rows []string
	for _, row := range in {
		jsonString, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		rows = append(rows, string(jsonString))
	}
	return rows, nil
}

func ResourcesToStringMap(resources []Resource, solution Solution) (out []map[string]interface{}) {
	version := time.Now().Unix()
	for _, resource := range resources {
		resourceMap := make(map[string]interface{})
		resourceMap["solutionName"] = solution.solutionName
		resourceMap["solutionNumber"] = solution.solutionNumber
		resourceMap["version"] = version
		resourceMap["resourceName"] = resource.resourceName
		resourceMap["resourceType"] = resource.resourceType
		attributes := make([]map[string]interface{}, 0)
		for attributeName, attributeValue := range resource.resourceAttributes {
			if attributeName != "depends_on" {
				attribute := make(map[string]interface{})
				attribute["name"] = attributeName
				attribute["value"] = attributeValue
				attributes = append(attributes, attribute)
			} else {
				resourceMap["dependsOn"] = attributeValue
			}
		}
		resourceMap["attributes"] = attributes
		out = append(out, resourceMap)
	}
	return
}

func DebugPrintPatternTable(matched []MatchedPattern) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Pattern", "Resource Count", "Weight", "Condition Count"})
	for i, pattern := range matched {
		t.AppendRow(table.Row{
			i,
			pattern.Pattern.PatternName,
			len(pattern.Resources),
			pattern.Pattern.Weight,
			pattern.ConditionCount,
		})
	}
	t.Render()
}

func PrintTextPatternTable(matched []MatchedPattern) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Pattern", "Target", "Resources"})
	for i, pattern := range matched {
		var resources = ""
		for _, resource := range pattern.Resources {
			resources = resources + resource.resourceType + "/" + resource.resourceName + ", "
		}
		t.AppendRow(table.Row{
			i,
			pattern.Pattern.PatternName,
			pattern.Pattern.Target,
			resources[:len(resources)-2],
		})
	}
	t.Render()
}

func PrintTextResourceTable(unmatched []string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Resource"})
	for i, resource := range unmatched {
		t.AppendRow(table.Row{
			i,
			resource,
		})
	}
	t.Render()
}
