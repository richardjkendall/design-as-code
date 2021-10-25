package main

import (
	"fmt"
	"sort"
)

type MatchedPattern struct {
	Pattern   Pattern
	Resources []Resource
}

func SolveForPriority(matches []MatchedPattern, resources []Resource) (solution []MatchedPattern, unmatched []string) {

	// create a map to track which resources have been used
	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	// sort by weight
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Pattern.Weight < matches[j].Pattern.Weight
	})

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

func SolvForMaxCoverage(matches []MatchedPattern, resources []Resource) (solution []MatchedPattern, unmatched []string) {

	// create a map to track which resources have been used
	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	// sort by number of resources per pattern
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i].Resources) > len(matches[j].Resources)
	})

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

func MatchPatternsToSolution(resources []Resource, patterns []Pattern) (matched []MatchedPattern, unmatched []string) {

	matchMap := make(map[string]bool)
	for _, resource := range resources {
		matchMap[resource.resourceType+"/"+resource.resourceName] = false
	}

	for _, pattern := range patterns {
		fmt.Printf("Attempting to match pattern: %s\n", pattern.PatternName)

		var mp MatchedPattern
		var mr []Resource

		mp.Pattern = pattern
		matchingRules := 0

		for _, rule := range pattern.Rules {
			fmt.Printf("\tWorking on rule for resource: %s\n", rule.Resource)
			matchingResources := 0

			for _, resource := range resources {
				match := true

				// is this rule for the resource we are looking at?
				if resource.resourceType == rule.Resource {

					// run through the conditions
					for _, condition := range rule.Conditions {
						fmt.Printf("\t\tChecking %s %s %s\n", condition.Attribute, condition.Operator, condition.Value)

						// does the the resource have the attributes the rule expects?
						_, present := resource.resourceAttributes[condition.Attribute]
						if present {
							// get the values into variables with shorter names
							expectedValue := condition.Value
							actualValue := resource.resourceAttributes[condition.Attribute]

							// check if the actual value matches the current value using the operator specified by the rule
							switch condition.Operator {
							case "eq":
								// check for equality
								if actualValue == expectedValue {
									match = SetTrueIfNotFalse(match)
									fmt.Printf("\t\t\tGot a match\n")
								} else {
									match = false
								}
							case "lt":
								// check for less than
								if actualValue < expectedValue {
									match = SetTrueIfNotFalse(match)
									fmt.Printf("\t\t\tGot a match\n")
								} else {
									match = false
								}
							case "gt":
								// check for greater than
								if actualValue > expectedValue {
									match = SetTrueIfNotFalse(match)
									fmt.Printf("\t\t\tGot a match\n")
								} else {
									match = false
								}
							// TODO: implement gte, lte, ne
							default:
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
			fmt.Printf("All rules for pattern %s matched\n", pattern.PatternName)
			mp.Resources = mr
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
