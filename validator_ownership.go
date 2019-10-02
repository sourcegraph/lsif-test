package main

import "fmt"

func (v *Validator) getOwnershipMap() (map[id]id, error) {
	if v.ownershipMap != nil {
		return v.ownershipMap, nil
	}

	ownershipMap := map[id]id{}
	err := v.forEachContainsEdge(func(line string, edge *edge1n) error {
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

func invertOwnershipMap(m map[id]id) map[id][]id {
	inverted := map[id][]id{}
	for k, v := range m {
		inverted[v] = append(inverted[v], k)
	}

	return inverted
}
