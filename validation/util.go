package validation

import (
	"github.com/sourcegraph/lsif-test/elements"
)

func (v *Validator) forEachVertex(label string, f func(line string, element *elements.Element) error) error {
	for _, line := range v.edges {
		edge, err := elements.ParseElement(line)
		if err != nil {
			return err
		}

		if edge.Label == label {
			if err := f(line, edge); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) forEachEdge(label string, f func(line string, edge *elements.Edge1n) error) error {
	for _, line := range v.edges {
		edge, err := elements.ParseEdge(line)
		if err != nil {
			return err
		}

		if edge.Label == label {
			if err := f(line, edge); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) forEachContainsEdge(f func(line string, edge *elements.Edge1n) error) error {
	return v.forEachEdge("contains", func(line string, edge *elements.Edge1n) error {
		parentElement, err := v.vertexElement(edge.OutV)
		if err != nil {
			return err
		}

		if parentElement.Label == "document" {
			return f(line, edge)
		}

		return nil
	})
}
