package main

import (
	"fmt"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type MatchedPattern struct {
	Pattern        Pattern
	ConditionCount int
	Resources      []Resource
}

func SolveForPriority(matches []MatchedPattern, resources []Resource) (solution []MatchedPattern, unmatched []string) {

	// create a map to track which resources have been used
	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println("\nMatched patterns before sorting:")
		DebugPrintPatternTable(matches)
	}

	// sort by specificity fist
	// biggest to smallest
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].ConditionCount > matches[j].ConditionCount
	})

	// sort by weight
	// smallest to biggest
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].Pattern.Weight < matches[j].Pattern.Weight
	})

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println("\nMatched patterns after sorting:")
		DebugPrintPatternTable(matches)
		fmt.Println("")
	}

	// got through the matches and select them till they are run out or we have covered all the resources
	for _, mp := range matches {
		// check that all resources are unused
		match := true
		for _, resource := range mp.Resources {
			if matchMap[resource.resourceType+"/"+resource.resourceName] {
				match = false
			}
		}
		if match {
			solution = append(solution, mp)
			// need to mark resources as used
			for _, resource := range mp.Resources {
				matchMap[resource.resourceType+"/"+resource.resourceName] = true
			}
		}
	}

	// populate unmatched
	for key, value := range matchMap {
		if !value {
			unmatched = append(unmatched, key)
		}
	}

	return
}

// TODO: need to solve for using the most specific matches first (in both solvers)
func SolvForMaxCoverage(matches []MatchedPattern, resources []Resource) (solution []MatchedPattern, unmatched []string) {

	// create a map to track which resources have been used
	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println("\nMatched patterns before sorting:")
		DebugPrintPatternTable(matches)
	}

	// sort by specificity fist
	// biggest to smallest
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].ConditionCount > matches[j].ConditionCount
	})

	// sort by number of resources per pattern
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i].Resources) > len(matches[j].Resources)
	})

	if log.GetLevel() >= log.DebugLevel {
		fmt.Println("\nMatched patterns after sorting:")
		DebugPrintPatternTable(matches)
		fmt.Println("")
	}

	// got through the matches and select them till they are run out or we have covered all the resources
	for _, mp := range matches {
		// check that all resources are unused
		match := true
		for _, resource := range mp.Resources {
			if matchMap[resource.resourceType+"/"+resource.resourceName] {
				match = false
			}
		}
		if match {
			solution = append(solution, mp)
			// need to mark resources as used
			for _, resource := range mp.Resources {
				matchMap[resource.resourceType+"/"+resource.resourceName] = true
			}
		}
	}

	// populate unmatched
	for key, value := range matchMap {
		if !value {
			unmatched = append(unmatched, key)
		}
	}

	return
}

func SetTrueIfNotFalse(in bool) bool {
	if in {
		return in
	}
	return false
}

func CheckRelation(actualValue interface{}, expectedValue string, operator string, expectedType string) bool {
	log.WithFields(log.Fields{
		"actual":   actualValue,
		"expected": expectedValue,
		"type":     expectedType,
		"operator": operator,
	}).Trace("Starting check relation")
	switch expectedType {
	case "string":
		actual := actualValue.(string)
		switch operator {
		case "eq":
			return actual == expectedValue
		case "lt":
			return actual < expectedValue
		case "gt":
			return actual > expectedValue
		default:
			log.Trace("No valid operator provided")
			return false
		}
	case "bool":
		actual := actualValue.(bool)
		expected := expectedValue == "true"
		return actual == expected
	case "int":
		actual := actualValue.(int)
		expected, err := strconv.Atoi(expectedValue)
		if err != nil {
			log.WithError(err).Error("Failed to convert string to integer")
			return false
		}
		switch operator {
		case "eq":
			return actual == expected
		case "lt":
			return actual < expected
		case "gt":
			return actual > expected
		default:
			log.Trace("No valid operator provided")
			return false
		}
	}
	log.Trace("No valid type provided")
	return false
}

func MatchPatternsToSolution(resources []Resource, patterns []Pattern, typemap map[string]map[string]string) (matched []MatchedPattern, unmatched []string) {

	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	for _, pattern := range patterns {
		log.WithFields(log.Fields{
			"pattern": pattern.PatternName,
		}).Debug("Attempting to match pattern")

		var mp MatchedPattern
		var mr []Resource

		mp.Pattern = pattern
		matchingRules := 0
		conditionCount := 0

		for _, rule := range pattern.Rules {
			log.WithFields(log.Fields{
				"pattern":  pattern.PatternName,
				"resource": rule.Resource,
			}).Debug("Working on rule for resource")
			matchingResources := 0

			for _, resource := range resources {
				match := true

				// is this rule for the resource we are looking at?
				if resource.resourceType == rule.Resource {

					// run through the conditions
					for _, condition := range rule.Conditions {
						log.WithFields(log.Fields{
							"pattern":       pattern.PatternName,
							"resource":      rule.Resource,
							"attribute":     condition.Attribute,
							"attributeType": typemap[resource.resourceType][condition.Attribute],
							"operator":      condition.Operator,
							"value":         condition.Value,
						}).Debug("Checking condition")

						// does the the resource have the attributes the rule expects?
						_, present := resource.resourceAttributes[condition.Attribute]
						if present {
							// get the values into variables with shorter names
							expectedValue := condition.Value
							actualValue := resource.resourceAttributes[condition.Attribute]

							// check if the actual value matches the current value using the operator specified by the rule
							// TODO: implement additional operators: gte, lte, link

							expectedType := typemap[resource.resourceType][condition.Attribute]
							if CheckRelation(actualValue, expectedValue, condition.Operator, expectedType) {
								log.Trace("Back from check relation with a +ve match")
								match = SetTrueIfNotFalse(match)
								conditionCount = conditionCount + 1
							} else {
								log.Trace("Back from check relation with a -ve match")
								match = false
							}

						} else {
							// attribute is not present for this resource, so this is an automatic no match
							match = false
						}
					}
				} else {
					match = false
				}

				// does the resource match?
				if match {
					matchingResources = matchingResources + 1
					mr = append(mr, resource)

					// update the matchmap to show which resources were matched
					matchMap[resource.resourceType+"/"+resource.resourceName] = true
				}

			}

			if matchingResources > 0 {
				matchingRules = matchingRules + 1
			}

		}

		// check to see if all the rules this pattern has were matched, if not all matched then this is a fail
		if matchingRules == len(pattern.Rules) {
			log.WithFields(log.Fields{
				"pattern": pattern.PatternName,
			}).Debug("All rules for pattern matched")
			mp.Resources = mr
			mp.ConditionCount = conditionCount
			matched = append(matched, mp)
		}

	}

	// create a list of unmatched resources which can be used by other consumers
	for key, value := range matchMap {
		if !value {
			unmatched = append(unmatched, key)
		}
	}

	return

}
