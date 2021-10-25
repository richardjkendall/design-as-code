package main

import "github.com/hashicorp/hcl/v2"

var solutionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "solution_name",
			Required: true,
		},
		{
			Name:     "apm_number",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
	},
}

var databaseSchema = hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "type",
			Required: true,
		},
		{
			Name:     "platform",
			Required: true,
		},
		{
			Name:     "arch",
			Required: false,
		},
		{
			Name:     "virtual",
			Required: false,
		},
		{
			Name:     "ha",
			Required: false,
		},
		{
			Name:     "role",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "sla",
		},
	},
}

var nasSchema = hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "type",
			Required: true,
		},
	},
}

var serverSchema = hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "os",
			Required: true,
		},
		{
			Name:     "virtual",
			Required: true,
		},
		{
			Name:     "hypervisor",
			Required: false,
		},
		{
			Name:     "arch",
			Required: false,
		},
		{
			Name:     "cores",
			Required: false,
		},
		{
			Name:     "memory",
			Required: false,
		},
		{
			Name:     "role",
			Required: false,
		},
		{
			Name:     "count",
			Required: false,
		},
		{
			Name:     "depends_on",
			Required: false,
		},
	},
}

var lbSchema = hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "protocol",
			Required: true,
		},
		{
			Name:     "backends",
			Required: true,
		},
	},
}

var schemaMap = map[string]hcl.BodySchema{
	"database":      databaseSchema,
	"server":        serverSchema,
	"nas":           nasSchema,
	"load_balancer": lbSchema,
}
