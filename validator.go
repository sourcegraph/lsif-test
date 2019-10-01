package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type lineValidator func(line string) error
type elementValidator func(line string, element *element) error

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
}

func NewValidator(schema *gojsonschema.Schema) *Validator {
	validator := &Validator{
		schema:   schema,
		vertices: map[id]string{},
		edges:    map[id]string{},
	}

	validator.elementValidators = map[string]elementValidator{
		"vertex": validator.validateVertex,
		"edge":   validator.validateEdge,
	}

	validator.vertexValidators = map[string]lineValidator{
		"metaData": validator.validateMetaDataVertex,
		"document": validator.validateDocumentVertex,
		"range":    validator.validateRangeVertex,
	}

	validator.edgeValidators = map[string]lineValidator{
		"contains":                validator.validateContainsEdge,
		"item":                    validator.validateItemEdge,
		"next":                    validator.validateEdge11([]string{"range", "resultSet"}, "resultSet"),
		"textDocument/definition": validator.validateEdge11([]string{"range", "resultSet"}, "definitionResult"),
		"textDocument/references": validator.validateEdge11([]string{"range", "resultSet"}, "referenceResult"),
		"textDocument/hover":      validator.validateEdge11([]string{"range", "resultSet"}, "hoverResult"),
		"moniker":                 validator.validateEdge11([]string{"range", "resultSet"}, "moniker"),
		"nextMoniker":             validator.validateEdge11([]string{"moniker"}, "moniker"),
		"packageInformation":      validator.validateEdge11([]string{"moniker"}, "packageInformation"),
	}

	return validator
}

func (v *Validator) ValidateLine(line string) error {
	if err := v.ensureMetadata(); err != nil {
		return err
	}

	v.lines++

	// TODO - enable/disable via flag
	// if err := v.validateSchema(line); err != nil {
	// 	return err
	// }

	element := &element{}
	if err := json.Unmarshal([]byte(line), &element); err != nil {
		return err
	}

	if err := v.elementValidators[element.Type](line, element); err != nil {
		return err
	}

	return nil
}

func (v *Validator) Process() error {
	// TODO - ensure all vertices reachable from a range
	// TODO - all ranges must belong to a document
	// TODO - ranges must not overlap
	// TODO - regression from other indexers

	fmt.Printf("%d vertices, %d edges\n", len(v.vertices), len(v.edges))
	return nil
}

//
// Element Validators

func (v *Validator) validateVertex(line string, element *element) error {
	if err := v.validate(v.vertexValidators, element.Label, line); err != nil {
		return err
	}

	if err := v.stashVertex(line, element.ID); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateEdge(line string, element *element) error {
	if err := v.validate(v.edgeValidators, element.Label, line); err != nil {
		return err
	}

	if err := v.stashEdge(line, element.ID); err != nil {
		return err
	}

	return nil
}

func (v *Validator) stashVertex(line string, id id) error {
	if _, ok := v.vertices[id]; ok {
		return fmt.Errorf("vertex %s already exists", id)
	}

	if _, ok := v.edges[id]; ok {
		return fmt.Errorf("vertex and edges cannot share id %s", id)
	}

	v.vertices[id] = line
	return nil
}

func (v *Validator) stashEdge(line string, id id) error {
	if _, ok := v.edges[id]; ok {
		return fmt.Errorf("edge %s already exists", id)
	}

	if _, ok := v.vertices[id]; ok {
		return fmt.Errorf("vertex and edges cannot share id %s", id)
	}

	v.edges[id] = line
	return nil
}

//
// Vertex Validators

func (v *Validator) validateMetaDataVertex(line string) error {
	metaData, err := parseMetaData(line)
	if err != nil {
		return err
	}

	url, err := url.Parse(metaData.ProjectRoot)
	if err != nil {
		return err
	}

	v.hasMetadata = true
	v.projectRoot = url
	return nil
}

func (v *Validator) validateDocumentVertex(line string) error {
	document, err := parseDocument(line)
	if err != nil {
		return err
	}

	url, err := url.Parse(document.URI)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(url.String(), v.projectRoot.String()) {
		return fmt.Errorf("document is not relative to project root")
	}

	return nil
}

func (v *Validator) validateRangeVertex(line string) error {
	documentRange, err := parseDocumentRange(line)
	if err != nil {
		return err
	}

	bounds := []int{
		documentRange.Start.Line,
		documentRange.Start.Character,
		documentRange.End.Line,
		documentRange.End.Character,
	}

	for _, bound := range bounds {
		if bound < 0 {
			return fmt.Errorf("illegal range bounds")
		}
	}

	if documentRange.Start.Line > documentRange.End.Line {
		return fmt.Errorf("illegal range extents")
	}

	if documentRange.Start.Line == documentRange.End.Line {
		if documentRange.Start.Character > documentRange.End.Character {
			return fmt.Errorf("illegal range extents")
		}
	}

	return nil
}

//
// Edge Validators

func (v *Validator) validateContainsEdge(line string) error {
	edge, err := parseEdge1n(line)
	if err != nil {
		return err
	}

	parentElement, err := v.vertexElement(edge.OutV)
	if err != nil {
		return err
	}

	if parentElement.Label == "document" {
		for _, inV := range edge.InVs {
			if err := v.ensureVertexType(inV, []string{"range"}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateItemEdge(line string) error {
	edge, err := parseEdge1n(line)
	if err != nil {
		return err
	}

	element, err := v.vertexElement(edge.OutV)
	if err != nil {
		return err
	}

	labels := []string{"range"}
	if element.Label == "referenceResult" {
		labels = append(labels, "referenceResult")
	}

	for _, inV := range edge.InVs {
		if err := v.ensureVertexType(inV, labels); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateEdge11(sources []string, result string) lineValidator {
	return func(line string) error {
		edge, err := parseEdge11(line)
		if err != nil {
			return err
		}

		if err := v.ensureVertexType(edge.OutV, sources); err != nil {
			return err
		}

		if err := v.ensureVertexType(edge.InV, []string{result}); err != nil {
			return err
		}

		return nil
	}
}

func (v *Validator) vertexElement(id id) (*element, error) {
	line, ok := v.vertices[id]
	if !ok {
		return nil, fmt.Errorf("no such vertex %s", id)
	}

	return parseElement(line)
}

func (v *Validator) ensureVertexType(id id, labels []string) error {
	element, err := v.vertexElement(id)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if element.Label == label {
			return nil
		}
	}

	return fmt.Errorf("expected vertex %s to be of type %s", id, strings.Join(labels, ", "))
}

//
// Common Helpers

func (v *Validator) ensureMetadata() error {
	if v.lines > 0 && !v.hasMetadata {
		return fmt.Errorf("metaData vertex must come first")
	}

	return nil
}

func (v *Validator) validateSchema(line string) error {
	result, err := v.schema.Validate(gojsonschema.NewStringLoader(line))
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("failed schema validation")
	}

	return nil
}

func (v *Validator) validate(validators map[string]lineValidator, label string, line string) error {
	if f, ok := validators[label]; ok {
		return f(line)
	}

	return nil
}
