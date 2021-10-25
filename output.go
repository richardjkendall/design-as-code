package main

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func PrintTextPatternTable(matched []MatchedPattern) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Pattern", "Target", "Resources"})
	for i, pattern := range matched {
		var resources = ""
		for _, resource := range pattern.Resources {
			resources = resource.resourceType + "/" + resource.resourceName + ", "
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