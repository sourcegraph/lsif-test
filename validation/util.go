package validation

import (
	"github.com/sourcegraph/lsif-test/elements"
)

func (v *Validator) forEachVertex(label string, f func(line string, element *elements.Element) bool) bool {
	allOk := true
	for _, line := range v.edges {
		edge, err := elements.ParseElement(line)
		if err != nil {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		if edge.Label == label {
			if !f(line, edge) {
				allOk = false
			}
		}
	}

	return allOk
}

func (v *Validator) forEachEdge(label string, f func(line string, edge *elements.Edge1n) bool) bool {
	allOk := true
	for _, line := range v.edges {
		edge, err := elements.ParseEdge(line)
		if err != nil {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		if edge.Label == label {
			if !f(line, edge) {
				allOk = false
			}
		}
	}

	return allOk
}

func (v *Validator) forEachContainsEdge(f func(line string, edge *elements.Edge1n) bool) bool {
	return v.forEachEdge("contains", func(line string, edge *elements.Edge1n) bool {
		parentElement, ok := v.vertexElement(edge.OutV)
		if !ok {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		if parentElement.Label == "document" {
			return f(line, edge)
		}

		return true
	})
}
