package validation

import (
	"fmt"

	"github.com/sourcegraph/lsif-test/elements"
)

func (v *Validator) getOwnershipMap() (map[elements.ID]elements.ID, bool) {
	if v.ownershipMap != nil {
		return v.ownershipMap, true
	}

	ownershipMap := map[elements.ID]elements.ID{}
	ok := v.forEachContainsEdge(func(line string, edge *elements.Edge1n) bool {
		for _, inV := range edge.InVs {
			if _, ok := ownershipMap[inV]; ok {
				// TODO - more context
				v.addError(ValidationError{Message: fmt.Sprintf("range %s claimed by multiple documents", inV)})
				return false
			}

			ownershipMap[inV] = edge.OutV
		}

		return true
	})

	if !ok {
		return nil, false
	}

	v.ownershipMap = ownershipMap
	return ownershipMap, true
}

func invertOwnershipMap(m map[elements.ID]elements.ID) map[elements.ID][]elements.ID {
	inverted := map[elements.ID][]elements.ID{}
	for k, v := range m {
		inverted[v] = append(inverted[v], k)
	}

	return inverted
}
