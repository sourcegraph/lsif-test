package validation

import (
	"github.com/sourcegraph/lsif-test/elements"
)

func (v *Validator) forEachVertex(label string, f func(lineContext LineContext) bool) bool {
	valid := true
	for _, lineContext := range v.vertices {
		if lineContext.Element.Label == label {
			if !f(lineContext) {
				valid = false
			}
		}
	}

	return valid
}

func (v *Validator) forEachEdge(label string, f func(lineContext LineContext, edge *elements.Edge1n) bool) bool {
	valid := true
	for _, lineContext := range v.edges {
		edge, err := elements.ParseEdge(lineContext.LineText)
		if err != nil {
			// already parsed
			panic("Unreachable!")
		}

		if edge.Label == label {
			if !f(lineContext, edge) {
				valid = false
			}
		}
	}

	return valid
}

func (v *Validator) forEachContainsEdge(f func(lineContext LineContext, edge *elements.Edge1n) bool) bool {
	return v.forEachEdge("contains", func(lineContext LineContext, edge *elements.Edge1n) bool {
		parentElement, ok := v.vertexElement(lineContext, edge.OutV)
		if !ok {
			// already parsed
			panic("Unreachable!")
		}

		if parentElement.Label == "document" {
			return f(lineContext, edge)
		}

		return true
	})
}
