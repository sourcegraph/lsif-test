package elements

import "encoding/json"

type Edge11 struct {
	*Element
	OutV ID `json:"outV"`
	InV  ID `json:"inV"`
}

type Edge1n struct {
	*Element
	OutV ID   `json:"outV"`
	InVs []ID `json:"inVs"`
}

func ParseEdge(line string) (*Edge1n, error) {
	edge1n := &Edge1n{}
	if err := json.Unmarshal([]byte(line), &edge1n); err != nil {
		return nil, err
	}

	if len(edge1n.InVs) == 0 {
		edge11 := &Edge11{}
		if err := json.Unmarshal([]byte(line), &edge11); err != nil {
			return nil, err
		}

		edge1n.InVs = append(edge1n.InVs, edge11.InV)
	}

	return edge1n, nil
}

func ParseEdge11(line string) (*Edge11, error) {
	edge := &Edge11{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func ParseEdge1n(line string) (*Edge1n, error) {
	edge := &Edge1n{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}
