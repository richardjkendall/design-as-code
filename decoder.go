package main

import "github.com/hashicorp/hcl/v2"

type Resource struct {
	resourceType       string
	resourceName       string
	resourceAttributes map[string]string
}

func DecodeBody(body *hcl.BodyContent, resourceType string, schemas map[string]hcl.BodySchema) ([]Resource, hcl.Diagnostics) {

	var resources []Resource

	ctx := &hcl.EvalContext{}
	for _, block := range body.Blocks.OfType("resource") {
		resourceType := block.Labels[0]

		var resource Resource
		attributes := make(map[string]string)

		resource.resourceName = block.Labels[1]
		resource.resourceType = resourceType

		schema := schemas[resourceType]
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
