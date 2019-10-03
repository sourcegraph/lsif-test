package validation

import (
	"fmt"

	"github.com/sourcegraph/lsif-test/elements"
)

func (v *Validator) getOwnershipMap() (map[elements.ID]elements.ID, error) {
	if v.ownershipMap != nil {
		return v.ownershipMap, nil
	}

	ownershipMap := map[elements.ID]elements.ID{}
	err := v.forEachContainsEdge(func(line string, edge *elements.Edge1n) error {
		for _, inV := range edge.InVs {
			if _, ok := ownershipMap[inV]; ok {
				return fmt.Errorf("range %s claimed by multiple documents", inV)
			}

			ownershipMap[inV] = edge.OutV
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	v.ownershipMap = ownershipMap
	return ownershipMap, nil
}

func invertOwnershipMap(m map[elements.ID]elements.ID) map[elements.ID][]elements.ID {
	inverted := map[elements.ID][]elements.ID{}
	for k, v := range m {
		inverted[v] = append(inverted[v], k)
	}

	return inverted
}
