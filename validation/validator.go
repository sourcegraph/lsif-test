package validation

import (
	"net/url"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

type Validator struct {
	schema            *gojsonschema.Schema
	disableJSONSchema bool
	errors            []ValidationError
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

type ElementValidator func(line string, element *elements.Element) bool
type LineValidator func(line string) bool

func NewValidator(schema *gojsonschema.Schema, disableJSONSchema bool) *Validator {
	validator := &Validator{
		schema:            schema,
		disableJSONSchema: disableJSONSchema,
		vertices:          map[elements.ID]string{},
		edges:             map[elements.ID]string{},
	}

	validator.elementValidators = validator.setupElementValidators()
	validator.vertexValidators = validator.setupVertexValidators()
	validator.edgeValidators = validator.setupEdgeValidators()
	return validator
}
