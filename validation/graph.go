package validation

import (
	"fmt"
	"sort"

	"github.com/sourcegraph/lsif-test/elements"
)

var whitelist = []string{"metaData", "project"}

func (v *Validator) ValidateGraph(stopOnError bool) bool {
	processors := []func() bool{
		v.ensureReachability,   // TODO - early out here as well
		v.ensureRangeOwnership, // TODO - early out here as well
		v.ensureDisjointRanges, // TODO - early out here as well
		v.ensureItemContains,   // TODO - early out here as well
	}

	allOk := true
	for _, f := range processors {
		if !f() {
			allOk = false
			if stopOnError {
				return false
			}
		}
	}

	fmt.Printf("%d vertices, %d edges\n", len(v.vertices), len(v.edges))
	return allOk
}

func (v *Validator) ensureReachability() bool {
	visited := map[elements.ID]bool{}
	ok := v.forEachContainsEdge(func(line string, edge *elements.Edge1n) bool {
		for _, inV := range append([]elements.ID{edge.OutV}, edge.InVs...) {
			visited[inV] = true
		}

		return true
	})

	if !ok {
		return false
	}

	changed := true
	for changed {
		changed = false

		for _, line := range v.edges {
			edge, err := elements.ParseEdge(line)
			if err != nil {
				// all lines have already been parsed
				panic("Unreachable!")
			}

			if _, ok := visited[edge.OutV]; ok {
				for _, inV := range edge.InVs {
					if _, ok := visited[inV]; !ok {
						changed = true
					}

					visited[inV] = true
				}
			}
		}
	}

	allOk := true

outer:
	for id, line := range v.vertices {
		element, err := elements.ParseElement(line)
		if err != nil {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		for _, label := range whitelist {
			if element.Label == label {
				continue outer
			}
		}

		if _, ok := visited[id]; !ok {
			allOk = false
			// TODO - more context
			v.addError(ValidationError{Message: fmt.Sprintf("unreachable vertex %s", id)})
		}
	}

	return allOk
}

func (v *Validator) ensureRangeOwnership() bool {
	ownedBy, ok := v.getOwnershipMap()
	if !ok {
		return false
	}

	return v.forEachVertex("range", func(line string, element *elements.Element) bool {
		if _, ok := ownedBy[element.ID]; !ok {
			// TODO - more context
			v.addError(ValidationError{Message: fmt.Sprintf("range %s not owned by any document", element.ID)})
			return false
		}

		return true
	})
}

func (v *Validator) ensureDisjointRanges() bool {
	ownershipMap, ok := v.getOwnershipMap()
	if !ok {
		return false
	}

	allOk := true
	for documentID, rangeIDs := range invertOwnershipMap(ownershipMap) {
		documentRanges := []*elements.DocumentRange{}
		for _, rangeID := range rangeIDs {
			documentRange, err := elements.ParseDocumentRange(v.vertices[rangeID])
			if err != nil {
				// all lines have already been parsed
				panic("Unreachable!")
			}

			documentRanges = append(documentRanges, documentRange)
		}

		if !v.ensureDisjoint(documentID, documentRanges) {
			allOk = false
		}
	}

	return allOk
}

func (v *Validator) ensureDisjoint(documentID elements.ID, documentRanges []*elements.DocumentRange) bool {
	sort.Slice(documentRanges, func(i, j int) bool {
		s1 := documentRanges[i].Start
		s2 := documentRanges[j].Start
		return s1.Line < s2.Line || (s1.Line == s2.Line && s1.Character < s2.Character)
	})

	allOk := true
	for i := 1; i < len(documentRanges); i++ {
		r1 := documentRanges[i-1]
		r2 := documentRanges[i]

		// TODO - can they share the same end point?
		if r1.End.Line > r2.Start.Line || (r1.End.Line == r2.Start.Line && r1.End.Character > r2.Start.Character) {
			allOk = false
			// TODO - more context
			v.addError(ValidationError{Message: fmt.Sprintf("ranges %s and %s overlap in document %s\n", r1.ID, r2.ID, documentID)})
		}
	}

	return allOk
}

func (v *Validator) ensureItemContains() bool {
	ownedBy, ok := v.getOwnershipMap()
	if !ok {
		return false
	}

	return v.forEachEdge("item", func(line string, edge *elements.Edge1n) bool {
		itemEdge, err := elements.ParseItemEdge(line)
		if err != nil {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		for _, inV := range edge.InVs {
			if ownedBy[inV] != itemEdge.Document {
				// TODO - more contexts
				v.addError(ValidationError{Message: fmt.Sprintf(
					"vertex %s not owned by document %s, as implied by item edge %s",
					inV,
					itemEdge.Document,
					edge.ID,
				)})

				return false
			}
		}

		return true
	})
}
