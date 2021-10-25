package main

import "github.com/hashicorp/hcl/v2/hclsimple"

type Patterns struct {
	SetName    string    `hcl:"pattern_set_name"`
	PatternSet []Pattern `hcl:"pattern,block"`
}

type Pattern struct {
	PatternName string `hcl:"pattern_name,label"`
	Description string `hcl:"description"`
	Weight      int    `hcl:"weight"`
	Target      string `hcl:"target"`
	Rules       []Rule `hcl:"rule,block"`
}

type Rule struct {
	Resource   string      `hcl:"resource"`
	Conditions []Condition `hcl:"condition,block"`
}

type Condition struct {
	Attribute string `hcl:"attribute"`
	Operator  string `hcl:"operator"`
	Value     string `hcl:"value"`
}

func LoadPatternLibrary(file string) (Patterns, error) {
	var patterns Patterns
	err := hclsimple.DecodeFile(file, nil, &patterns)
	if err != nil {
		return patterns, err
	}
	return patterns, nil
}
