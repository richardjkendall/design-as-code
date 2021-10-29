package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"

	log "github.com/sirupsen/logrus"
)

type Resource struct {
	resourceType       string
	resourceName       string
	resourceAttributes map[string]interface{}
}

func ExpressionToValue(expr hcl.Expression, ctx *hcl.EvalContext, variableName string, schema map[string]string) (interface{}, hcl.Diagnostics) {
	log.WithFields(log.Fields{
		"variableName": variableName,
		"variableType": schema[variableName],
	}).Trace("Expression to value conversion starting")

	// if this is a depends_on, we need to handle it seperately
	if variableName == "depends_on" {
		log.Trace("This is depends_on, so a list of strings")
		var out []string
		diag := gohcl.DecodeExpression(expr, ctx, &out)
		if diag != nil && diag.HasErrors() {
			return nil, diag
		}
		log.WithFields(log.Fields{
			"variableName": variableName,
			"value":        out,
		}).Trace("Got a list of strings")
		return out, nil
	}

	switch schema[variableName] {
	case "string":
		var out string
		diag := gohcl.DecodeExpression(expr, ctx, &out)
		if diag != nil && diag.HasErrors() {
			return nil, diag
		}
		log.WithFields(log.Fields{
			"variableName": variableName,
			"value":        out,
		}).Trace("Got a string")
		return out, nil
	case "bool":
		var out bool
		diag := gohcl.DecodeExpression(expr, ctx, &out)
		if diag != nil && diag.HasErrors() {
			return nil, diag
		}
		log.WithFields(log.Fields{
			"variableName": variableName,
			"value":        out,
		}).Trace("Got a bool")
		return out, nil
	case "int":
		var out int
		diag := gohcl.DecodeExpression(expr, ctx, &out)
		if diag != nil && diag.HasErrors() {
			return nil, diag
		}
		log.WithFields(log.Fields{
			"variableName": variableName,
			"value":        out,
		}).Trace("Got an int")
		return out, nil
	default:
		log.WithFields(log.Fields{
			"variableName": variableName,
		}).Trace("Got no value")
		return nil, nil
	}
}

func DecodeBody(body *hcl.BodyContent, resourceType string, schemas map[string]hcl.BodySchema, typemap map[string]map[string]string) ([]Resource, hcl.Diagnostics) {

	var resources []Resource

	variables := make(map[string][]string)

	// first pass to populate the parsing context
	for _, block := range body.Blocks.OfType("resource") {
		resourceType := block.Labels[0]
		resourceName := block.Labels[1]
		_, present := variables[resourceType]
		if present {
			variables[resourceType] = append(variables[resourceType], resourceName)
		} else {
			variables[resourceType] = append(make([]string, 0), resourceName)
		}
	}

	objectMapForEval := make(map[string]cty.Value)
	for resource, instances := range variables {
		variableMap := make(map[string]cty.Value)
		for _, instance := range instances {
			variableMap[instance] = cty.StringVal(resource + "." + instance)
		}
		objectMapForEval[resource] = cty.ObjectVal(variableMap)
	}

	ctx := &hcl.EvalContext{
		Variables: objectMapForEval,
	}
	// second pass to get all the attributes
	for _, block := range body.Blocks.OfType("resource") {
		resourceType := block.Labels[0]

		var resource Resource
		attributes := make(map[string]interface{})

		resource.resourceName = block.Labels[1]
		resource.resourceType = resourceType

		schema := schemas[resourceType]
		contents, diagnostics := block.Body.Content(&schema)
		if diagnostics != nil && diagnostics.HasErrors() {
			return nil, diagnostics
		}

		for _, attribute := range contents.Attributes {
			val, diag := ExpressionToValue(attribute.Expr, ctx, attribute.Name, typemap[resourceType])
			if diag != nil && diag.HasErrors() {
				return nil, diag
			}
			log.WithFields(log.Fields{
				"value": val,
			}).Trace("Got a value back from ExpressionToValue")

			attributes[attribute.Name] = val
		}

		resource.resourceAttributes = attributes

		resources = append(resources, resource)
	}

	return resources, nil
}
