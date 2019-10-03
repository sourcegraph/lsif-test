package validation

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

func (v *Validator) ValidateLine(line string) bool {
	ok := true
	if v.lines > 0 && !v.hasMetadata {
		ok = false
		// TODO - more context
		v.addError(ValidationError{Message: "metaData vertex must occur before any other element"})
	}

	v.lines++

	if !v.disableJSONSchema {
		if !v.validateSchema(line) {
			return false
		}
	}

	element, err := elements.ParseElement(line)
	if err != nil {
		v.addLineError(line, "failed to parse element")
		return false
	}

	if !v.elementValidators[element.Type](line, element) {
		ok = false
	}

	return ok
}

//
// Element Validators

func (v *Validator) setupElementValidators() map[string]ElementValidator {
	return map[string]ElementValidator{
		"vertex": v.validateVertex,
		"edge":   v.validateEdge,
	}
}

func (v *Validator) validateVertex(line string, element *elements.Element) bool {
	ok1 := v.validate(v.vertexValidators, element.Label, line)
	ok2 := v.stashVertex(line, element.ID)
	return ok1 && ok2
}

func (v *Validator) validateEdge(line string, element *elements.Element) bool {
	ok1 := v.validate(v.edgeValidators, element.Label, line)
	ok2 := v.stashEdge(line, element.ID)
	return ok1 && ok2
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

func (v *Validator) validateMetaDataVertex(line string) bool {
	if v.hasMetadata {
		v.addLineError(line, "metadata vertex defined multiple times")
		return false
	}

	metaData, err := elements.ParseMetaData(line)
	if err != nil {
		v.addLineError(line, "failed to parse metadata element")
		return false
	}

	url, err := url.Parse(metaData.ProjectRoot)
	if err != nil {
		v.addLineError(line, "project root is not a valid URL")
		return false
	}

	v.hasMetadata = true
	v.projectRoot = url
	return true
}

func (v *Validator) validateDocumentVertex(line string) bool {
	document, err := elements.ParseDocument(line)
	if err != nil {
		v.addLineError(line, "failed to parse document element")
		return false
	}

	url, err := url.Parse(document.URI)
	if err != nil {
		v.addLineError(line, "document uri is not a valid URL")
		return false
	}

	if v.projectRoot != nil && !strings.HasPrefix(url.String(), v.projectRoot.String()) {
		v.addLineError(line, "document is not relative to project root")
		return false
	}

	return true
}

func (v *Validator) validateRangeVertex(line string) bool {
	documentRange, err := elements.ParseDocumentRange(line)
	if err != nil {
		v.addLineError(line, "failed to parse range")
		return false
	}

	bounds := []int{
		documentRange.Start.Line,
		documentRange.Start.Character,
		documentRange.End.Line,
		documentRange.End.Character,
	}

	for _, bound := range bounds {
		if bound < 0 {
			v.addLineError(line, "illegal range bounds")
			return false
		}
	}

	if documentRange.Start.Line > documentRange.End.Line {
		v.addLineError(line, "illegal range extents")
		return false
	}

	if documentRange.Start.Line == documentRange.End.Line {
		if documentRange.Start.Character > documentRange.End.Character {
			v.addLineError(line, "illegal range extents")
			return false
		}
	}

	return true
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

func (v *Validator) validateContainsEdge(line string) bool {
	edge, err := elements.ParseEdge1n(line)
	if err != nil {
		v.addLineError(line, "failed to parse edge")
		return false
	}

	if len(edge.InVs) == 0 {
		v.addLineError(line, "inVs is an empty list")
		return false
	}

	parentElement, ok := v.vertexElement(edge.OutV)
	if !ok {
		return false
	}

	if parentElement.Label == "document" {
		for _, inV := range edge.InVs {
			if !v.ensureVertexType(inV, []string{"range"}) {
				return false
			}
		}
	}

	return true
}

func (v *Validator) validateItemEdge(line string) bool {
	edge, err := elements.ParseItemEdge(line)
	if err != nil {
		v.addLineError(line, "failed to parse item edge")
		return false
	}

	if len(edge.InVs) == 0 {
		v.addLineError(line, "inVs is an empty list")
		return false
	}

	element, ok := v.vertexElement(edge.OutV)
	if !ok {
		return false
	}

	labels := []string{"range"}
	if element.Label == "referenceResult" {
		labels = append(labels, "referenceResult")
	}

	if !v.ensureVertexType(edge.Document, []string{"document"}) {
		return false
	}

	for _, inV := range edge.InVs {
		if !v.ensureVertexType(inV, labels) {
			return false
		}
	}

	return true
}

func (v *Validator) validateEdge11(sources []string, result string) LineValidator {
	return func(line string) bool {
		edge, err := elements.ParseEdge11(line)
		if err != nil {
			v.addLineError(line, "failed to parse edge")
			return false
		}

		if !v.ensureVertexType(edge.OutV, sources) {
			return false
		}

		if !v.ensureVertexType(edge.InV, []string{result}) {
			return false
		}

		return true
	}
}

//
// Helpers

func (v *Validator) validate(validators map[string]LineValidator, label string, line string) bool {
	if f, ok := validators[label]; ok {
		return f(line)
	}

	return true
}

func (v *Validator) validateSchema(line string) bool {
	result, err := v.schema.Validate(gojsonschema.NewStringLoader(line))
	if err != nil {
		v.addLineError(line, "failed schema validation")
		return false
	}

	if !result.Valid() {
		// TODO - better messages here
		v.addLineError(line, "failed schema validation")
		return false
	}

	return true
}

func (v *Validator) vertexElement(id elements.ID) (*elements.Element, bool) {
	line, ok := v.vertices[id]
	if !ok {
		// TODO - more context
		v.addError(ValidationError{Message: fmt.Sprintf("no such vertex %s", id)})
		return nil, false
	}

	element, err := elements.ParseElement(line)
	if err != nil {
		// TODO - more context
		v.addError(ValidationError{Message: fmt.Sprintf("failed to parse vertex with ID %s", id)})
		return nil, false
	}

	return element, true
}

func (v *Validator) ensureVertexType(id elements.ID, labels []string) bool {
	element, ok := v.vertexElement(id)
	if !ok {
		return false
	}

	for _, label := range labels {
		if element.Label == label {
			return true
		}
	}

	// TODO - more context
	v.addError(ValidationError{Message: fmt.Sprintf("expected vertex %s to be of type %s", id, strings.Join(labels, ", "))})
	return false
}

func (v *Validator) stashVertex(line string, id elements.ID) bool {
	if _, ok := v.vertices[id]; ok {
		v.addLineError(line, fmt.Sprintf("vertex %s already exists", id))
		return false
	}

	if _, ok := v.edges[id]; ok {
		v.addLineError(line, fmt.Sprintf("vertex and edges cannot share id %s", id))
		return false
	}

	v.vertices[id] = line
	return true
}

func (v *Validator) stashEdge(line string, id elements.ID) bool {
	if _, ok := v.edges[id]; ok {
		v.addLineError(line, fmt.Sprintf("edge %s already exists", id))
		return false
	}

	if _, ok := v.vertices[id]; ok {
		v.addLineError(line, fmt.Sprintf("vertex and edges cannot share id %s", id))
		return false
	}

	v.edges[id] = line
	return true
}
