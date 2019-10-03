package validation

import (
	"net/url"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

type LineValidator func(line string) error
type ElementValidator func(line string, element *elements.Element) error

type Validator struct {
	schema            *gojsonschema.Schema
	elementValidators map[string]ElementValidator
	vertexValidators  map[string]LineValidator
	edgeValidators    map[string]LineValidator
	vertices          map[elements.ID]string
	edges             map[elements.ID]string
	hasMetadata       bool
	projectRoot       *url.URL
	lines             int
	ownershipMap      map[elements.ID]elements.ID
}

func NewValidator(schema *gojsonschema.Schema) *Validator {
	validator := &Validator{
		schema:   schema,
		vertices: map[elements.ID]string{},
		edges:    map[elements.ID]string{},
	}

	validator.elementValidators = validator.setupElementValidators()
	validator.vertexValidators = validator.setupVertexValidators()
	validator.edgeValidators = validator.setupEdgeValidators()
	return validator
}
