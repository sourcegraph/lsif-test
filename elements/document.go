package elements

import "encoding/json"

type Document struct {
	*Element
	URI string `json:"uri"`
}

func ParseDocument(line string) (*Document, error) {
	document := &Document{}
	if err := json.Unmarshal([]byte(line), &document); err != nil {
		return nil, err
	}

	return document, nil
}
