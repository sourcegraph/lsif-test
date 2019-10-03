package elements

import "encoding/json"

type Element struct {
	ID    ID     `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

func ParseElement(line string) (*Element, error) {
	element := &Element{}
	if err := json.Unmarshal([]byte(line), &element); err != nil {
		return nil, err
	}

	return element, nil
}
