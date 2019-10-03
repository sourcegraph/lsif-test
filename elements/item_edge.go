package elements

import "encoding/json"

type ItemEdge struct {
	*Element
	OutV     ID   `json:"outV"`
	InVs     []ID `json:"inVs"`
	Document ID   `json:"document"`
}

func ParseItemEdge(line string) (*ItemEdge, error) {
	edge := &ItemEdge{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}
