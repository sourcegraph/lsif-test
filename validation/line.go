package validation

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/lsif-test/elements"
	"github.com/xeipuuv/gojsonschema"
)

func (v *Validator) ValidateLine(line string) bool {
	defer func() { v.lines++ }()

	valid := true
	if v.lines == 1 && !v.hasMetadata {
		valid = false
		v.addError("metaData vertex must occur before any other element").At(line, v.lines)
	}

	if !v.disableJSONSchema {
		result, err := v.schema.Validate(gojsonschema.NewStringLoader(line))
		if err != nil {
			v.addError("failed schema validation").At(line, v.lines)
			return false
		}

		if !result.Valid() {
			// TODO - get message from result
			v.addError("failed schema validation").At(line, v.lines)
			return false
		}
	}

	element, err := elements.ParseElement(line)
	if err != nil {
		v.addError("failed to parse element").At(line, v.lines)
		return false
	}

	lineContext := LineContext{
		Element:   element,
		LineText:  line,
		LineIndex: v.lines,
	}

	return !v.elementValidators[element.Type](lineContext) && valid
}

//
// Element Validators

func (v *Validator) setupElementValidators() map[string]ValidatorFunc {
	return map[string]ValidatorFunc{
		"vertex": v.validateVertex,
		"edge":   v.validateEdge,
	}
}

func (v *Validator) validateVertex(lineContext LineContext) bool {
	ok1 := v.validate(v.vertexValidators, lineContext)
	ok2 := v.stashVertex(lineContext)
	return ok1 && ok2
}

func (v *Validator) validateEdge(lineContext LineContext) bool {
	ok1 := v.validate(v.edgeValidators, lineContext)
	ok2 := v.stashEdge(lineContext)
	return ok1 && ok2
}

//
// Vertex Validators

func (v *Validator) setupVertexValidators() map[string]ValidatorFunc {
	return map[string]ValidatorFunc{
		"metaData": v.validateMetaDataVertex,
		"document": v.validateDocumentVertex,
		"range":    v.validateRangeVertex,
	}
}

func (v *Validator) validateMetaDataVertex(lineContext LineContext) bool {
	if v.hasMetadata {
		v.addError("metadata vertex defined multiple times").Link(lineContext)
		return false
	}

	metaData, err := elements.ParseMetaData(lineContext.LineText)
	if err != nil {
		v.addError("failed to parse metadata element").Link(lineContext)
		return false
	}

	url, err := url.Parse(metaData.ProjectRoot)
	if err != nil {
		v.addError("project root is not a valid URL").Link(lineContext)
		return false
	}

	v.hasMetadata = true
	v.projectRoot = url
	return true
}

func (v *Validator) validateDocumentVertex(lineContext LineContext) bool {
	document, err := elements.ParseDocument(lineContext.LineText)
	if err != nil {
		v.addError("failed to parse document element").Link(lineContext)
		return false
	}

	url, err := url.Parse(document.URI)
	if err != nil {
		v.addError("document uri is not a valid URL").Link(lineContext)
		return false
	}

	if v.projectRoot != nil && !strings.HasPrefix(url.String(), v.projectRoot.String()) {
		v.addError("document is not relative to project root").Link(lineContext)
		return false
	}

	return true
}

func (v *Validator) validateRangeVertex(lineContext LineContext) bool {
	documentRange, err := elements.ParseDocumentRange(lineContext.LineText)
	if err != nil {
		v.addError("failed to parse range").Link(lineContext)
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
			v.addError("illegal range bounds").Link(lineContext)
			return false
		}
	}

	if documentRange.Start.Line > documentRange.End.Line {
		v.addError("illegal range extents").Link(lineContext)
		return false
	}

	if documentRange.Start.Line == documentRange.End.Line {
		if documentRange.Start.Character > documentRange.End.Character {
			v.addError("illegal range extents").Link(lineContext)
			return false
		}
	}

	return true
}

//
// Edge Validators

func (v *Validator) setupEdgeValidators() map[string]ValidatorFunc {
	return map[string]ValidatorFunc{
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

func (v *Validator) validateContainsEdge(lineContext LineContext) bool {
	edge, err := elements.ParseEdge1n(lineContext.LineText)
	if err != nil {
		v.addError("failed to parse edge").Link(lineContext)
		return false
	}

	if len(edge.InVs) == 0 {
		v.addError("inVs is an empty list").Link(lineContext)
		return false
	}

	parentElement, ok := v.vertexElement(lineContext, edge.OutV)
	if !ok {
		return false
	}

	if parentElement.Label == "document" {
		for _, inV := range edge.InVs {
			if !v.ensureVertexType(lineContext, inV, []string{"range"}) {
				return false
			}
		}
	}

	return true
}

func (v *Validator) validateItemEdge(lineContext LineContext) bool {
	edge, err := elements.ParseItemEdge(lineContext.LineText)
	if err != nil {
		v.addError("failed to parse item edge").Link(lineContext)
		return false
	}

	if len(edge.InVs) == 0 {
		v.addError("inVs is an empty list").Link(lineContext)
		return false
	}

	element, ok := v.vertexElement(lineContext, edge.OutV)
	if !ok {
		return false
	}

	labels := []string{"range"}
	if element.Label == "referenceResult" {
		labels = append(labels, "referenceResult")
	}

	if !v.ensureVertexType(lineContext, edge.Document, []string{"document"}) {
		return false
	}

	for _, inV := range edge.InVs {
		if !v.ensureVertexType(lineContext, inV, labels) {
			return false
		}
	}

	return true
}

func (v *Validator) validateEdge11(sources []string, result string) ValidatorFunc {
	return func(lineContext LineContext) bool {
		edge, err := elements.ParseEdge11(lineContext.LineText)
		if err != nil {
			v.addError("failed to parse edge").Link(lineContext)
			return false
		}

		if !v.ensureVertexType(lineContext, edge.OutV, sources) {
			return false
		}

		if !v.ensureVertexType(lineContext, edge.InV, []string{result}) {
			return false
		}

		return true
	}
}

//
// Helpers

func (v *Validator) validate(validators map[string]ValidatorFunc, lineContext LineContext) bool {
	if f, ok := validators[lineContext.Element.Label]; ok {
		return f(lineContext)
	}

	return true
}

func (v *Validator) vertexElement(parentLineContext LineContext, id elements.ID) (*elements.Element, bool) {
	lineContext, ok := v.vertices[id]
	if !ok {
		v.addError("no such vertex %s", id).Link(parentLineContext)
		return nil, false
	}

	return lineContext.Element, true
}

func (v *Validator) ensureVertexType(parentLineContext LineContext, id elements.ID, labels []string) bool {
	lineContext, ok := v.vertices[id]
	if !ok {
		v.addError("no such vertex %s", id).Link(parentLineContext)
		return false
	}

	for _, label := range labels {
		if lineContext.Element.Label == label {
			return true
		}
	}

	v.addError("expected vertex %s to be of type %s", id, strings.Join(labels, ", ")).Link(lineContext, parentLineContext)
	return false
}

func (v *Validator) stashVertex(lineContext LineContext) bool {
	return v.stash(lineContext, v.vertices, v.edges, "vertex")
}

func (v *Validator) stashEdge(lineContext LineContext) bool {
	return v.stash(lineContext, v.edges, v.vertices, "edge")
}

func (v *Validator) stash(lineContext LineContext, m1, m2 map[elements.ID]LineContext, elementType string) bool {
	if _, ok := m1[lineContext.Element.ID]; ok {
		v.addError("%s %s already exists", elementType, lineContext.Element.ID).Link(lineContext, m1[lineContext.Element.ID])
		return false
	}

	if _, ok := m2[lineContext.Element.ID]; ok {
		v.addError("vertex and edges cannot share id %s", lineContext.Element.ID).Link(lineContext, m2[lineContext.Element.ID])
		return false
	}

	m1[lineContext.Element.ID] = lineContext
	return true
}
