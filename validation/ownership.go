package validation

import (
	"github.com/sourcegraph/lsif-test/elements"
)

type ownershipContext struct {
	outV        elements.ID
	lineContext LineContext
}

func (v *Validator) getOwnershipMap() (map[elements.ID]ownershipContext, bool) {
	if v.ownershipMap != nil {
		return v.ownershipMap, true
	}

	ownershipMap := map[elements.ID]ownershipContext{}
	valid := v.forEachContainsEdge(func(lineContext LineContext, edge *elements.Edge1n) bool {
		valid := true
		for _, inV := range edge.InVs {
			if previousOwner, ok := ownershipMap[inV]; ok {
				v.addError("range %s already claimed by document %s", inV, previousOwner.outV).Link(
					lineContext,
					previousOwner.lineContext,
				)

				valid = false
				continue
			}

			ownershipMap[inV] = ownershipContext{
				outV:        edge.OutV,
				lineContext: lineContext,
			}
		}

		return valid
	})

	if !valid {
		return nil, false
	}

	v.ownershipMap = ownershipMap
	return ownershipMap, true
}

func invertOwnershipMap(m map[elements.ID]ownershipContext) map[elements.ID][]elements.ID {
	inverted := map[elements.ID][]elements.ID{}
	for k, v := range m {
		inverted[v.outV] = append(inverted[v.outV], k)
	}

	return inverted
}
