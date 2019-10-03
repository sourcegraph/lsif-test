package validation

import (
	"net/url"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

type Validator struct {
	schema            *gojsonschema.Schema
	disableJSONSchema bool
	errors            []*ValidationError
	elementValidators map[string]ValidatorFunc
	vertexValidators  map[string]ValidatorFunc
	edgeValidators    map[string]ValidatorFunc
	vertices          map[elements.ID]LineContext
	edges             map[elements.ID]LineContext
	hasMetadata       bool
	projectRoot       *url.URL
	lines             int
	ownershipMap      map[elements.ID]ownershipContext
}

type LineContext struct {
	Element   *elements.Element
	LineText  string
	LineIndex int
}

type ValidatorFunc func(lineContext LineContext) bool

func NewValidator(schema *gojsonschema.Schema, disableJSONSchema bool) *Validator {
	validator := &Validator{
		schema:            schema,
		disableJSONSchema: disableJSONSchema,
		vertices:          map[elements.ID]LineContext{},
		edges:             map[elements.ID]LineContext{},
	}

	validator.elementValidators = validator.setupElementValidators()
	validator.vertexValidators = validator.setupVertexValidators()
	validator.edgeValidators = validator.setupEdgeValidators()
	return validator
}

func (v *Validator) Stats() (int, int) {
	return len(v.vertices), len(v.edges)
}
