package main

import (
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	log "github.com/sirupsen/logrus"
)

// convertValueToString recurses through cty.Value structures and converts them to a string representation
func convertValueToString(val cty.Value) string {
	log.WithFields(log.Fields{
		"fieldType": val.Type().FriendlyName(),
	}).Trace("convertValueToString")

	// for basic types we can return the string representation right away
	if val.Type() == cty.String {
		return val.AsString()
	}
	if val.Type() == cty.Number {
		return numberToString(val)
	}
	if val.Type() == cty.Bool {
		return boolToString(val)
	}

	// for tuples (which seems to include lists) we need to iterate
	if val.Type().IsTupleType() {
		var ret []string
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			ret = append(ret, convertValueToString(v))
		}
		return "[" + strings.Join(ret, ", ") + "]"
	}

	// for objects (which seems to include maps) we need to iterate though the attributes
	// need to sort attributes first so that comparisons will work
	if val.Type().IsObjectType() {
		var ret []string
		atys := val.Type().AttributeTypes()
		attributeNames := make([]string, 0, len(atys))
		for name := range atys {
			attributeNames = append(attributeNames, name)
		}
		sort.Strings(attributeNames)
		for _, name := range attributeNames {
			ret = append(ret, name+"="+convertValueToString(val.GetAttr(name)))
		}
		return "{" + strings.Join(ret, ", ") + "}"
	}

	if val.Type().HasDynamicTypes() {
		log.Trace("Field has dyanmic types")

	}

	// if we get here we have an issue
	return "ERROR: cannot convert!"
}

// numberToString converts cty.Value number to a string representation
func numberToString(val cty.Value) string {
	return val.AsBigFloat().String()
}

// boolToString converts cty.Value bool to a string representation
func boolToString(val cty.Value) string {
	var ret bool
	gocty.FromCtyValue(val, &ret)
	if ret {
		return "true"
	}
	return "false"
}

/*
// trimAll takes the elements of a slice of strings and trims all the whitespace off the strings in the slice
func trimAll(input []string) []string {
	output := make([]string, len(input))
	for i, s := range input {
		output[i] = strings.Trim(s, " \r\n")
	}
	return output
}*/
