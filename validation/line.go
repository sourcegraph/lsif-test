package validation

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

func (v *Validator) ValidateLine(line string) error {
	if err := v.ensureMetadata(); err != nil {
		return err
	}

	v.lines++

	if !v.disableJSONSchema {
		if err := v.validateSchema(line); err != nil {
			return err
		}
	}

	element, err := elements.ParseElement(line)
	if err != nil {
		return err
	}

	if err := v.elementValidators[element.Type](line, element); err != nil {
		return err
	}

	return nil
}

//
// Element Validators

func (v *Validator) setupElementValidators() map[string]ElementValidator {
	return map[string]ElementValidator{
		"vertex": v.validateVertex,
		"edge":   v.validateEdge,
	}
}

func (v *Validator) validateVertex(line string, element *elements.Element) error {
	if err := v.validate(v.vertexValidators, element.Label, line); err != nil {
		return err
	}

	if err := v.stashVertex(line, element.ID); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateEdge(line string, element *elements.Element) error {
	if err := v.validate(v.edgeValidators, element.Label, line); err != nil {
		return err
	}

	if err := v.stashEdge(line, element.ID); err != nil {
		return err
	}

	return nil
}

//
// Vertex Validators

func (v *Validator) setupVertexValidators() map[string]LineValidator {
	return map[string]LineValidator{
		"metaData": v.validateMetaDataVertex,
		"document": v.validateDocumentVertex,
		"range":    v.validateRangeVertex,
	}
}

func (v *Validator) validateMetaDataVertex(line string) error {
	if v.hasMetadata {
		return fmt.Errorf("metadata vertex defined multiple times")
	}

	metaData, err := elements.ParseMetaData(line)
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
	document, err := elements.ParseDocument(line)
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
	documentRange, err := elements.ParseDocumentRange(line)
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

func (v *Validator) setupEdgeValidators() map[string]LineValidator {
	return map[string]LineValidator{
		"contains":                v.validateContainsEdge,
		"item":                    v.validateItemEdge,
		"next":                    v.validateEdge11([]string{"range", "resultSet"}, "resultSet"),
		"textDocument/definition": v.validateEdge11([]string{"range", "resultSet"}, "definitionResult"),
		"textDocument/references": v.validateEdge11([]string{"range", "resultSet"}, "referenceResult"),
		"textDocument/hover":      v.validateEdge11([]string{"range", "resultSet"}, "hoverResult"),
		"moniker":                 v.validateEdge11([]string{"range", "resultSet"}, "moniker"),
		"nextMoniker":             v.validateEdge11([]string{"moniker"}, "moniker"),
		"packageInformation":      v.validateEdge11([]string{"moniker"}, "packageInformation"),
	}
}

func (v *Validator) validateContainsEdge(line string) error {
	edge, err := elements.ParseEdge1n(line)
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
	edge, err := elements.ParseItemEdge(line)
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

	if err := v.ensureVertexType(edge.Document, []string{"document"}); err != nil {
		return err
	}

	for _, inV := range edge.InVs {
		if err := v.ensureVertexType(inV, labels); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateEdge11(sources []string, result string) LineValidator {
	return func(line string) error {
		edge, err := elements.ParseEdge11(line)
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

//
// Helpers

func (v *Validator) validate(validators map[string]LineValidator, label string, line string) error {
	if f, ok := validators[label]; ok {
		return f(line)
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

func (v *Validator) vertexElement(id elements.ID) (*elements.Element, error) {
	line, ok := v.vertices[id]
	if !ok {
		return nil, fmt.Errorf("no such vertex %s", id)
	}

	return elements.ParseElement(line)
}

func (v *Validator) ensureVertexType(id elements.ID, labels []string) error {
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

func (v *Validator) ensureMetadata() error {
	if v.lines > 0 && !v.hasMetadata {
		return fmt.Errorf("metaData vertex must come first")
	}

	return nil
}

func (v *Validator) stashVertex(line string, id elements.ID) error {
	if _, ok := v.vertices[id]; ok {
		return fmt.Errorf("vertex %s already exists", id)
	}

	if _, ok := v.edges[id]; ok {
		return fmt.Errorf("vertex and edges cannot share id %s", id)
	}

	v.vertices[id] = line
	return nil
}

func (v *Validator) stashEdge(line string, id elements.ID) error {
	if _, ok := v.edges[id]; ok {
		return fmt.Errorf("edge %s already exists", id)
	}

	if _, ok := v.vertices[id]; ok {
		return fmt.Errorf("vertex and edges cannot share id %s", id)
	}

	v.edges[id] = line
	return nil
}
