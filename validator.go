package main

import (
	"net/url"

	"github.com/xeipuuv/gojsonschema"
)

type lineValidator func(line string) error
type elementValidator func(line string, element *Element) error

type Validator struct {
	schema            *gojsonschema.Schema
	elementValidators map[string]elementValidator
	vertexValidators  map[string]lineValidator
	edgeValidators    map[string]lineValidator
	vertices          map[id]string
	edges             map[id]string
	hasMetadata       bool
	projectRoot       *url.URL
	lines             int
	ownershipMap      map[id]id
}

func NewValidator(schema *gojsonschema.Schema) *Validator {
	validator := &Validator{
		schema:   schema,
		vertices: map[id]string{},
		edges:    map[id]string{},
	}

	validator.elementValidators = validator.setupElementValidators()
	validator.vertexValidators = validator.setupVertexValidators()
	validator.edgeValidators = validator.setupEdgeValidators()
	return validator
}
