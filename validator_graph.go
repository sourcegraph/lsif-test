package main

import (
	"fmt"
	"sort"
)

var whitelist = []string{"metaData", "project"}

func (v *Validator) ValidateGraph() error {
	processors := []func() error{
		v.ensureReachability,
		v.ensureRangeOwnership,
		v.ensureDisjointRanges,
		v.ensureItemContains,
	}

	for _, f := range processors {
		if err := f(); err != nil {
			return err
		}
	}

	fmt.Printf("%d vertices, %d edges\n", len(v.vertices), len(v.edges))
	return nil
}

func (v *Validator) ensureReachability() error {
	visited := map[id]bool{}

	err := v.forEachContainsEdge(func(line string, edge *edge1n) error {
		for _, inV := range append([]id{edge.OutV}, edge.InVs...) {
			visited[inV] = true
		}

		return nil
	})

	if err != nil {
		return err
	}

	changed := true
	for changed {
		changed = false

		for _, line := range v.edges {
			edge, err := parseEdge(line)
			if err != nil {
				return err
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

outer:
	for id, line := range v.vertices {
		element, err := parseElement(line)
		if err != nil {
			return err
		}

		for _, label := range whitelist {
			if element.Label == label {
				continue outer
			}
		}

		if _, ok := visited[id]; !ok {
			return fmt.Errorf("unreachable vertex %s", id)
		}
	}

	return nil
}

func (v *Validator) ensureRangeOwnership() error {
	ownedBy, err := v.getOwnershipMap()
	if err != nil {
		return err
	}

	return v.forEachVertex("range", func(line string, element *Element) error {
		if _, ok := ownedBy[element.ID]; !ok {
			return fmt.Errorf("range %s not owned by any document", element.ID)
		}

		return nil
	})
}

func (v *Validator) ensureDisjointRanges() error {
	ownershipMap, err := v.getOwnershipMap()
	if err != nil {
		return err
	}

	for documentID, rangeIDs := range invertOwnershipMap(ownershipMap) {
		documentRanges := []*documentRange{}
		for _, rangeID := range rangeIDs {
			documentRange, err := parseDocumentRange(v.vertices[rangeID])
			if err != nil {
				return err
			}

			documentRanges = append(documentRanges, documentRange)
		}

		if err := v.ensureDisjoint(documentID, documentRanges); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) ensureDisjoint(documentID id, documentRanges []*documentRange) error {
	sort.Slice(documentRanges, func(i, j int) bool {
		s1 := documentRanges[i].Start
		s2 := documentRanges[j].Start
		return s1.Line < s2.Line || (s1.Line == s2.Line && s1.Character < s2.Character)
	})

	for i := 1; i < len(documentRanges); i++ {
		r1 := documentRanges[i-1]
		r2 := documentRanges[i]

		// TODO - can they share the same end point?
		if r1.End.Line > r2.Start.Line || (r1.End.Line == r2.Start.Line && r1.End.Character > r2.Start.Character) {
			fmt.Printf("%#v\n", r1)
			fmt.Printf("%#v\n", r2)
			fmt.Printf("ranges %s and %s overlap in document %s\n", r1.ID, r2.ID, documentID)
		}
	}

	return nil
}

func (v *Validator) ensureItemContains() error {
	ownedBy, err := v.getOwnershipMap()
	if err != nil {
		return err
	}

	return v.forEachEdge("item", func(line string, edge *edge1n) error {
		itemEdge, err := parseItemEdge(line)
		if err != nil {
			return err
		}

		for _, inV := range edge.InVs {
			if ownedBy[inV] != itemEdge.Document {
				return fmt.Errorf(
					"vertex %s not owned by document %s, as implied by item edge %s",
					inV,
					itemEdge.Document,
					edge.ID,
				)
			}
		}

		return nil
	})
}
