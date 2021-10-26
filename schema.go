package main

import (
	"errors"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func ReadSchema(schemas map[string]hcl.BodySchema, typemap map[string]map[string]string) error {
	log.Debug("Reading solution-spec.yml")
	schema, err := ioutil.ReadFile("solution-spec.yml")
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Parsing schema...")
	data := make(map[interface{}]interface{})
	err = yaml.Unmarshal(schema, &data)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Creating HCL schema typemap and map of schemas")
	for k, v := range data {
		log.WithFields(log.Fields{
			"resource": k,
		}).Debug("Got resource block from spec")
		typemap[k.(string)] = make(map[string]string)
		attributes := []hcl.AttributeSchema{}
		blocks := []hcl.BlockHeaderSchema{}
		for vname, vtype := range v.(map[string]interface{}) {
			log.WithFields(log.Fields{
				"resource":     k,
				"variableName": vname,
				"variableType": vtype,
			}).Debug("Got variable in resource")
			if vname == "depends_on" {
				return errors.New("do not specify 'depends_on' as an attribute to a block")
			}
			typemap[k.(string)][vname] = vtype.(string)
			if vtype.(string) == "block" {
				block := hcl.BlockHeaderSchema{
					Type: vname,
				}
				blocks = append(blocks, block)
			} else {
				attribute := hcl.AttributeSchema{
					Name:     vname,
					Required: false,
				}
				attributes = append(attributes, attribute)
			}
		}
		// add depends_on
		depends_on := hcl.AttributeSchema{
			Name:     "depends_on",
			Required: false,
		}
		attributes = append(attributes, depends_on)
		schema := hcl.BodySchema{
			Attributes: attributes,
			Blocks:     blocks,
		}
		schemas[k.(string)] = schema
	}

	return nil

}

var solutionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "solution_name",
			Required: true,
		},
		{
			Name:     "solution_number",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
	},
}
