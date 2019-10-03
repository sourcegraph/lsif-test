package validation

import (
	"sort"

	"github.com/sourcegraph/lsif-test/elements"
)

var whitelist = []string{"metaData", "project"}

func (v *Validator) ValidateGraph(stopOnError bool) bool {
	// TODO - obey stopOnError for these functions a swell
	processors := []func() bool{
		v.ensureReachability,
		v.ensureRangeOwnership,
		v.ensureDisjointRanges,
		v.ensureItemContains,
	}

	valid := true
	for _, f := range processors {
		if !f() {
			valid = false
			if stopOnError {
				return false
			}
		}
	}

	return valid
}

func (v *Validator) ensureReachability() bool {
	visited := map[elements.ID]bool{}
	v.forEachContainsEdge(func(lineContext LineContext, edge *elements.Edge1n) bool {
		for _, inV := range append([]elements.ID{edge.OutV}, edge.InVs...) {
			visited[inV] = true
		}

		return true
	})

	changed := true
	for changed {
		changed = false

		for _, lineContext := range v.edges {
			edge, err := elements.ParseEdge(lineContext.LineText)
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

	valid := true

outer:
	for id, lineContext := range v.vertices {
		for _, label := range whitelist {
			if lineContext.Element.Label == label {
				continue outer
			}
		}

		if _, ok := visited[id]; !ok {
			valid = false
			v.addError("vertex %s unreachable from any range", id).Link(v.vertices[id])
		}
	}

	return valid
}

func (v *Validator) ensureRangeOwnership() bool {
	ownedBy, ok := v.getOwnershipMap()
	if !ok {
		return false
	}

	return v.forEachVertex("range", func(lineContext LineContext) bool {
		if _, ok := ownedBy[lineContext.Element.ID]; !ok {
			v.addError("range %s not owned by any document", lineContext.Element.ID).Link(lineContext)
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

	valid := true
	for documentID, rangeIDs := range invertOwnershipMap(ownershipMap) {
		documentRanges := []*elements.DocumentRange{}
		for _, rangeID := range rangeIDs {
			documentRange, err := elements.ParseDocumentRange(v.vertices[rangeID].LineText)
			if err != nil {
				// all lines have already been parsed
				panic("Unreachable!")
			}

			documentRanges = append(documentRanges, documentRange)
		}

		if !v.ensureDisjoint(documentID, documentRanges) {
			valid = false
		}
	}

	return valid
}

func (v *Validator) ensureDisjoint(documentID elements.ID, documentRanges []*elements.DocumentRange) bool {
	sort.Slice(documentRanges, func(i, j int) bool {
		s1 := documentRanges[i].Start
		s2 := documentRanges[j].Start
		return s1.Line < s2.Line || (s1.Line == s2.Line && s1.Character < s2.Character)
	})

	valid := true
	for i := 1; i < len(documentRanges); i++ {
		r1 := documentRanges[i-1]
		r2 := documentRanges[i]

		// TODO - can they touch?
		if r1.End.Line > r2.Start.Line || (r1.End.Line == r2.Start.Line && r1.End.Character > r2.Start.Character) {
			valid = false
			v.addError("ranges overlap").Link(v.vertices[r1.ID], v.vertices[r2.ID])
		}
	}

	return valid
}

func (v *Validator) ensureItemContains() bool {
	ownedBy, ok := v.getOwnershipMap()
	if !ok {
		return false
	}

	return v.forEachEdge("item", func(lineContext LineContext, edge *elements.Edge1n) bool {
		itemEdge, err := elements.ParseItemEdge(lineContext.LineText)
		if err != nil {
			// all lines have already been parsed
			panic("Unreachable!")
		}

		valid := true
		for _, inV := range edge.InVs {
			if ownedBy[inV].outV != itemEdge.Document {
				valid = false
				v.addError("vertex %s not owned by document %s, as implied by item edge %s",
					inV,
					itemEdge.Document,
					edge.ID,
				).Link(lineContext, ownedBy[inV].lineContext)
			}
		}

		return valid
	})
}
