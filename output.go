package main

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

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
